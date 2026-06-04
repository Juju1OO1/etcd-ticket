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
	if err := mq.Publish(ctx, orderID, td.UserName, td.PhoneNum, td.Area, "success"); err != nil {
		return err
	}
	fmt.Printf("[ORDER] %s order published to Redis Stream  orderID=%s\n", td.UserName, orderID)
	return nil
}

// StartOrderWorker 在 main.go 啟動時以 goroutine 執行，持續消費 Redis Stream
func StartOrderWorker(ctx context.Context) {
	go func() {
		fmt.Println("[WORKER] started, waiting for messages...")
		for {
			msgs, err := mq.Consume(ctx)
			if err != nil {
				if ctx.Err() != nil {
					fmt.Println("[WORKER] received shutdown signal, stopping")
					return
				}
				fmt.Printf("[WORKER] consume failed: %v\n", err)
				continue
			}

			for _, msg := range msgs {
				fmt.Printf("[WORKER] received msgID=%s\n", msg.ID)
				if err := processMessage(ctx, msg.ID, msg.Values); err != nil {
					fmt.Printf("[WORKER] ✗ process failed msgID=%s err=%v\n", msg.ID, err)
					// 不 Ack，訊息留在 pending list，下次重新消費
					continue
				}
				// 寫入成功才 Ack
				if err := mq.Ack(ctx, msg.ID); err != nil {
					fmt.Printf("[WORKER] ✗ ack failed msgID=%s err=%v\n", msg.ID, err)
					continue
				}
				fmt.Printf("[WORKER] ✓ acknowledged msgID=%s\n", msg.ID)
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

	if err := repository.InsertOrder(ctx, order); err != nil {
		return err
	}
	fmt.Printf("[WORKER] ✓ inserted into PostgreSQL  user=%s area=%d orderID=%s\n", order.UserName, order.Area, order.OrderID)
	return nil
}
