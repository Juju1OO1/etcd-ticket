# cmd/server

程式入口點，負責依序初始化所有服務並啟動。

## main.go

### 啟動順序

1. 讀取 `configs/config.yaml`，取得 PostgreSQL DSN 和 Redis addr
2. `db.Init()` — 建立 PostgreSQL 連線池，確認連得到 DB
3. `mq.Init()` — 連接 Redis，建立 consumer group `order-service`
4. `service.StartOrderWorker()` — 啟動背景 goroutine，開始監聽 Redis Stream
5. `service.InitEtcd()` — 寫入各區域的初始票數與狀態到 etcd
6. HTTP server 啟動（成員 3/4 負責）

### 注意

- 所有 `Init()` 都是 singleton，只會執行一次
- `defer db.Close()` / `defer mq.Close()` / `defer etcd.Close()` 確保程式正常關閉時釋放連線
- `select {}` 讓 main goroutine 不退出，等待 HTTP server 接管
