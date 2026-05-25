# internal/mq

Redis Stream 訊息佇列，負責訂單訊息的發布、消費與確認。

---

## redis_stream.go

### 常數
| 常數 | 值 | 說明 |
|------|-----|------|
| StreamKey | `order:stream` | Redis Stream 的 key 名稱 |
| ConsumerGroup | `order-service` | 消費者群組名稱 |
| ConsumerName | `worker-1` | 消費者名稱 |

### Init(ctx, addr)
main.go 啟動時呼叫一次：
- 建立 Redis 連線（singleton）
- Ping 確認連線成功
- 建立 consumer group `order-service`（已存在則忽略）

### Publish(ctx, orderID, userName, phoneNum, area, status)
用 `XADD` 把訂單資料寫進 Redis Stream，訊息格式：
```
order_id:  "uuid-xxxx"
user_name: "henry"
phone_num: "0912345678"
area:      1
status:    "success"
```
訊息寫入後持久化保存，程式重啟不會遺失。

### Consume(ctx)
用 `XREADGROUP` 從 Stream 讀取最多 10 筆未消費訊息，`Block: 0` 代表沒有新訊息時永久阻塞等待，不做 polling。

### Ack(ctx, msgID)
用 `XACK` 確認訊息已處理完成，Redis 才會從 pending list 移除該訊息。
若不呼叫 Ack，訊息會留在 pending list，下次 Consume 時會重新拿到，確保不丟單。

### Get()
供其他套件取得 Redis client 實例。

### Close()
程式結束時釋放 Redis 連線，於 main.go 的 defer 呼叫。
