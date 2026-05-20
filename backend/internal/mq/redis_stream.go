package mq

import (
	"context"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
)

const (
	StreamKey     = "order:stream"
	ConsumerGroup = "order-service"
	ConsumerName  = "worker-1"
)

var (
	once   sync.Once
	rdb    *redis.Client
)

// Init 在 main.go 啟動時呼叫一次
func Init(ctx context.Context, addr string) error {
	var initErr error
	once.Do(func() {
		rdb = redis.NewClient(&redis.Options{Addr: addr})
		if err := rdb.Ping(ctx).Err(); err != nil {
			initErr = fmt.Errorf("redis 連線失敗: %w", err)
			return
		}
		// 建立 consumer group，若已存在則忽略
		err := rdb.XGroupCreateMkStream(ctx, StreamKey, ConsumerGroup, "0").Err()
		if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
			initErr = fmt.Errorf("建立 consumer group 失敗: %w", err)
			return
		}
		fmt.Println("redis connected:", addr)
	})
	return initErr
}

// Get 供其他套件取得 redis client
func Get() *redis.Client {
	if rdb == nil {
		panic("redis 尚未初始化，請先呼叫 mq.Init()")
	}
	return rdb
}

// Close 在程式結束時釋放連線
func Close() {
	if rdb != nil {
		rdb.Close()
	}
}

// Publish 把訂單事件寫進 Redis Stream
func Publish(ctx context.Context, orderID, userName, phoneNum string, area int, status string) error {
	return Get().XAdd(ctx, &redis.XAddArgs{
		Stream: StreamKey,
		Values: map[string]interface{}{
			"order_id":  orderID,
			"user_name": userName,
			"phone_num": phoneNum,
			"area":      area,
			"status":    status,
		},
	}).Err()
}

// Consume 從 Redis Stream 讀取一批訊息（blocking）
func Consume(ctx context.Context) ([]redis.XMessage, error) {
	streams, err := Get().XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    ConsumerGroup,
		Consumer: ConsumerName,
		Streams:  []string{StreamKey, ">"},
		Count:    10,
		Block:    0, // 0 代表永久阻塞等待新訊息
	}).Result()
	if err != nil {
		return nil, err
	}
	if len(streams) == 0 {
		return nil, nil
	}
	return streams[0].Messages, nil
}

// Ack 確認訊息已消費完成，Redis 才會從 pending list 移除
func Ack(ctx context.Context, msgID string) error {
	return Get().XAck(ctx, StreamKey, ConsumerGroup, msgID).Err()
}
