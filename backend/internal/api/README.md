# internal/api

HTTP API 層，包含路由設定、middleware 與各功能 handler。

---

## router.go

### NewRouter(rl RateLimitConfig)

組裝整個 Gin router，回傳可直接 `Run()` 的 `*gin.Engine`。

掛載順序：
1. `gin.Logger()` / `gin.Recovery()` — 基本 log 與 panic 保護
2. `CORSMiddleware()` — 允許前端跨域呼叫
3. `GET /healthz` — 探活，不掛 rate limit
4. `RateLimitMiddleware()` — 套用到所有 `/api/*` 路由

### 路由表

| Method | Path | Handler | 說明 |
|--------|------|---------|------|
| GET | `/healthz` | HealthzHandler | 探活 + 查看累計被擋次數 |
| POST | `/api/tickets/reserve` | ReserveHandler | 第一階段搶票 |
| POST | `/api/tickets/checkout` | CheckoutHandler | 第二階段結帳 |
| GET | `/api/tickets/status` | StatusHandler | 查詢各區剩餘票數與狀態 |
| GET | `/api/orders` | OrdersHandler | 查詢用戶購票紀錄 |

---

## handler.go

搶票相關 handler。

### ReserveHandler — POST /api/tickets/reserve

解析 `user_name`、`phone_num`、`area`，呼叫 `service.Lock_And_Hold()` 競爭 etcd 鎖並佔位。
成功回傳 `reserved: true`；無票或已被佔回傳 409。

### CheckoutHandler — POST /api/tickets/checkout

呼叫 `service.Checkout()` 執行 CAS 扣票，再呼叫 `service.PublishOrder()` 把訂單丟進 Redis Stream。
成功回傳 `checked_out: true, order_published: true`。

### StatusHandler — GET /api/tickets/status

支援 query param `?area=1&area=2` 或 `?areas=1,2`，查詢各區剩餘票數與 etcd 狀態（on/off）。

---

## order.go

### OrdersHandler — GET /api/orders?user_name=xxx

查詢指定 `user_name` 的所有購票紀錄，呼叫 `repository.GetOrdersByUser()`，依建立時間降冪回傳。
`user_name` 為必填，空值回傳 400。

---

## middleware.go

### IPRateLimiter

Per-IP token bucket 限流器，使用 `sync.Map` 存各 IP 的 `*rate.Limiter`。
背景 goroutine 每 3 分鐘清掉 10 分鐘內無活動的 IP，避免記憶體洩漏。

設定由 `config.yaml` 傳入（`rps`、`burst`），目前預設 RPS=20、Burst=40。

### RateLimitMiddleware(l)

Per-IP 限流 gin middleware，超過限制回傳 429 並附 `Retry-After: 1` header。

### CORSMiddleware()

允許所有 origin 的跨域請求（`GET`、`POST`、`OPTIONS`），適用於前端 Vite dev server 場景。

### HealthzHandler(l)

`GET /healthz` 探活端點，回傳 `status: ok` 與累計被擋次數（`blocked_total`），方便壓測時觀察。

---

## ws.go

預留檔，WebSocket 由獨立的 `wsserver`（port 8888）負責，API Gateway 不實作 WS handler。
