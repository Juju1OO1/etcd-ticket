package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"etcd-ticket/internal/etcd"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// 第一次初始化時，會在etcd裡面建立好每個區域的 current_limit、limit、status 等 key，並設定好初始值
func GetAvailableTickets(ctx context.Context, areaID int) (int64, error) {
	area := "area" + strconv.Itoa(areaID)
	client := etcd.New()

	currentLimitKey := area + "/limit/current_limit"
	holderPrefix := area + "/holder/"
	payingPrefix := area + "/paying/"

	// 1. 先讀 current_limit
	curLimitStr, exist, err := client.Get(ctx, currentLimitKey)
	if err != nil {
		return 0, err
	}
	if !exist {
		return 0, fmt.Errorf("%s 不存在", currentLimitKey)
	}

	//2. 將curLimStr轉成int
	currentLimit, err := strconv.ParseInt(curLimitStr, 10, 64)
	if err != nil {
		return 0, err
	}

	//3. 讀取holder
	holderCount, err := client.CountPrefix(ctx, holderPrefix)
	if err != nil {
		return 0, err
	}

	//4. 讀取paying
	payingCount, err := client.CountPrefix(ctx, payingPrefix)
	if err != nil {
		return 0, err
	}

	available := currentLimit - holderCount - payingCount
	if available < 0 {
		available = 0
	}

	return available, nil
}

// 一直監聽，只要發生變化就會重新算一次剩餘票數
func WatchAvailableTickets(ctx context.Context, areaID int) (<-chan int64, error) {
	area := "area" + strconv.Itoa(areaID)
	client := etcd.New()

	channel := make(chan int64)
	
	// 派生新 goroutine 來監聽 etcd 的變化事件，並把剩餘票數送到 channel 裡，這樣就不會阻塞主流程。
	go func() {
		defer close(channel)

		sendAvailable := func() bool {
			available, err := GetAvailableTickets(ctx, areaID) // 這個方法會透過 RPC 請求去 etcd 讀取該區域目前的「總上限」、「被佔用數」、「付款中數」
			if err != nil {
				return false
			}
			fmt.Printf("[WATCHER] area%d available -> %d\n", areaID, available)
			select {
			case channel <- available:
				return true
			case <-ctx.Done():
				return false
			}
		}

		if !sendAvailable() {
			return // 初始值發送失敗，可能是 context 已經取消了
		}

		//用watch去監聽不同的prefix
		watchCurrentLimit := client.Cli.Watch(ctx, area+"/limit/current_limit")
		watchHolder := client.Cli.Watch(ctx, area+"/holder/", clientv3.WithPrefix())
		watchPaying := client.Cli.Watch(ctx, area+"/paying/", clientv3.WithPrefix())

		// 這是一個阻塞型的事件循環（Event Loop）。當 etcd 沒有任何資料變動、且外部沒有關機時，這個背景工人會卡在這一行進入不耗 CPU 算力的休眠狀態。
		for {
			select {
			case resp, ok := <-watchCurrentLimit:
				if !ok {
					return
				}
				if err := resp.Err(); err != nil {
					return
				}
				if !sendAvailable() {
					return
				}

			case resp, ok := <-watchHolder:
				if !ok {
					return
				}
				if err := resp.Err(); err != nil {
					return
				}
				if !sendAvailable() {
					return
				}

			case resp, ok := <-watchPaying:
				if !ok {
					return
				}
				if err := resp.Err(); err != nil {
					return
				}
				if !sendAvailable() {
					return
				}

			case <-ctx.Done():
				return
			}
		}
	}()
	return channel, nil
}

type SoldTicketEvent struct {
	UserName string
	Phone    string
	AreaID   int
}

func WatchSoldTickets(ctx context.Context, areaID int) (<-chan SoldTicketEvent, error) {

	// 除錯
	fmt.Printf(
		"WatchSoldTickets Start area=%d\n",
		areaID)
	// 除錯

	area := "area" + strconv.Itoa(areaID)
	client := etcd.New()

	soldPrefix := area + "/sold/"
	channel := make(chan SoldTicketEvent)

	go func() {
		defer close(channel)

		watchSold := client.Cli.Watch(ctx, soldPrefix, clientv3.WithPrefix())

		for {
			select {
			case resp, ok := <-watchSold:
				if !ok {
					return
				}
				if err := resp.Err(); err != nil {
					return
				}

				for _, event := range resp.Events {
					fmt.Printf(
						"Sold Event Receive key=%s value=%s\n",
						string(event.Kv.Key),
						string(event.Kv.Value),
					)

					if event.Type != clientv3.EventTypePut {
						continue
					}

					key := string(event.Kv.Key)
					userName := strings.TrimPrefix(key, soldPrefix)
					phone := string(event.Kv.Value)
					fmt.Printf("[WATCHER] area%d new sale: %s / %s\n", areaID, userName, phone)

					select {
					case channel <- SoldTicketEvent{
						UserName: userName,
						Phone:    phone,
						AreaID:   areaID,
					}:
					case <-ctx.Done():
						return
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return channel, nil
}
