# cmd/server

程式入口點，負責依序初始化所有服務並啟動。

## main.go

### 啟動順序

1. 讀取 `configs/config.yaml`，取得 PostgreSQL DSN、Redis addr、Rate Limit 設定
2. `db.Init()` — 建立 PostgreSQL 連線池，確認連得到 DB
3. `mq.Init()` — 連接 Redis，建立 consumer group `order-service`
4. `service.StartOrderWorker()` — 啟動背景 goroutine，開始監聽 Redis Stream
5. `service.InitEtcd()` — 寫入各區域的初始票數與狀態到 etcd
6. `wsserver.Start()` — 啟動 WebSocket server（port 8888），推送即時票況給前端
7. `watcher.StartHTTPClient()` — 監聽 etcd 剩餘票數變化，透過 WebSocket 廣播
8. `watcher.StartSoldTicketHTTPClient()` — 監聽 etcd 成交紀錄，透過 WebSocket 廣播
9. `api.NewRouter()` — 啟動 HTTP server（port 8080）

### API 路由

| Method | Path | 說明 |
|--------|------|------|
| POST | `/api/tickets/reserve` | 第一階段搶票（etcd 佔位） |
| POST | `/api/tickets/checkout` | 第二階段結帳（CAS 扣票 + 寫 Redis Stream） |
| GET | `/api/tickets/status` | 查詢各區剩餘票數與售票狀態 |
| GET | `/api/orders` | 查詢指定用戶的購票紀錄（`?user_name=xxx`） |

### 啟動指令

```bash
cd backend
go run cmd/server/main.go
```

> **注意**：若 port 8080 或 8888 已被佔用（例如上次未正常結束），請先用 `taskkill //F //PID <PID>` 或 `fg` + Ctrl+C 結束舊 process 再重啟。

### 注意事項

- 所有 `Init()` 都是 singleton，只會執行一次
- `defer db.Close()` / `defer mq.Close()` / `defer etcd.Close()` 確保程式正常關閉時釋放連線
- 建議不用 `&` 背景執行，改開獨立 terminal，Ctrl+C 才能正確終止所有 goroutine
