package service

import (
	"context"
	"etcd-ticket/internal/model"
	"etcd-ticket/internal/mq"
	"etcd-ticket/internal/repository"
	"fmt"
	"strconv"

	"github.com/google/uuid"
)

// PublishOrder 供成員4的 handler 在 Checkout() 成功後呼叫
func PublishOrder(ctx context.Context, td TicketData) error {
	orderID := uuid.NewString()
	return mq.Publish(ctx, orderID, td.UserName, td.PhoneNum, td.Area, "success")
}

// StartOrderWorker 在 main.go 啟動時以 goroutine 執行，持續消費 Redis Stream
func StartOrderWorker(ctx context.Context) {
	go func() {
		fmt.Println("[order worker] 啟動，等待訂單訊息...")
		for {
			select {
			case <-ctx.Done():
				fmt.Println("[order worker] 收到關閉信號，停止")
				return
			default:
				msgs, err := mq.Consume(ctx)
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					fmt.Printf("[order worker] consume 失敗: %v\n", err)
					continue
				}

				for _, msg := range msgs {
					if err := processMessage(ctx, msg.ID, msg.Values); err != nil {
						fmt.Printf("[order worker] 處理訊息失敗 msgID=%s err=%v\n", msg.ID, err)
						// 不 Ack，訊息留在 pending list，下次重新消費
						continue
					}
					// 寫入成功才 Ack
					if err := mq.Ack(ctx, msg.ID); err != nil {
						fmt.Printf("[order worker] ack 失敗 msgID=%s err=%v\n", msg.ID, err)
					}
				}
			}
		}
	}()
}

func processMessage(ctx context.Context, msgID string, values map[string]interface{}) error {
	area, err := strconv.Atoi(fmt.Sprintf("%v", values["area"]))
	if err != nil {
		return fmt.Errorf("解析 area 失敗: %w", err)
	}

	order := model.Order{
		OrderID:  fmt.Sprintf("%v", values["order_id"]),
		UserName: fmt.Sprintf("%v", values["user_name"]),
		PhoneNum: fmt.Sprintf("%v", values["phone_num"]),
		Area:     area,
		Status:   fmt.Sprintf("%v", values["status"]),
	}

	return repository.InsertOrder(ctx, order)
}
