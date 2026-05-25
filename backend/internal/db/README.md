# internal/db

PostgreSQL 連線管理與資料庫初始化設定。

## postgres.go

### Init(ctx, dsn)

main.go 啟動時呼叫一次，建立 `pgxpool` 連線池（singleton）並 Ping 確認連線。
DSN 格式：`postgres://user:password@host:port/dbname`

### Get()

供其他套件（repository）取得連線池，若未初始化則 panic。

### Close()

程式結束時釋放連線池，於 main.go 的 defer 呼叫。

---

## schema.sql

定義 `orders` 資料表，Docker 第一次啟動時自動執行。

### orders 表欄位


| 欄位       | 型別         | 說明                                          |
| ---------- | ------------ | --------------------------------------------- |
| id         | BIGSERIAL    | 自動遞增主鍵                                  |
| order_id   | VARCHAR(36)  | UUID，唯一訂單識別碼                          |
| user_name  | VARCHAR(100) | 購票用戶名稱                                  |
| phone_num  | VARCHAR(20)  | 購票用戶電話                                  |
| area       | INT          | 購票區域編號                                  |
| status     | VARCHAR(20)  | 訂單狀態（success / cancelled），預設 success |
| created_at | TIMESTAMPTZ  | 建立時間，DB 自動填入                         |

索引：

- `idx_orders_user` — 依 user_name 查詢
- `idx_orders_area` — 依 area 查詢

---

## docker-compose.yaml

一鍵啟動 PostgreSQL 和 Redis 容器。


| 服務     | Image       | Port                          |
| -------- | ----------- | ----------------------------- |
| postgres | postgres:16 | 5433（對外）→ 5432（容器內） |
| redis    | redis:7     | 6379                          |

PostgreSQL 使用 5433 對外，避免與本機預設 5432 衝突。
schema.sql 掛載進 `/docker-entrypoint-initdb.d/`，第一次啟動自動建表。

### 啟動指令

```bash
cd backend/internal/db
docker compose up -d
```

> **注意**：Redis 已包含在 docker-compose 裡（`etcd-ticket-redis`），不要另外 `docker run redis`，否則會造成 port 衝突。

---

## 啟動步驟

### 第一次啟動

### 前置條件

- 已安裝 Go、Docker、PostgreSQL（本機）

### 步驟一：啟動 Redis

```powershell
docker run -d --name redis-test -p 6379:6379 redis:alpine

```bash
cd backend/internal/db
docker compose up -d
```

### 重置（完整清除後重新建立）

如果之前已經跑過，要從頭重現：

```bash
cd backend/internal/db
docker compose down -v
docker compose up -d
```

`-v` 會刪除 volume，讓 schema.sql 下次啟動時重新執行。

### 確認容器正常運作

```bash
docker ps
```

應看到 `etcd-ticket-postgres`（port 5433）和 `etcd-ticket-redis`（port 6379）都是 `Up` 狀態。

---

## 本地測試：驗證資料寫入流程

測試完整流程：`PublishOrder` 發訊息到 Redis Stream → `StartOrderWorker` 消費 → 寫入 PostgreSQL。

### 前置條件
- 已安裝 Go、Docker
- 容器已啟動（見上方啟動步驟）

### 步驟一：執行測試

```bash
cd backend
go run test/order/main.go
```

預期輸出：

```
postgres connected
redis connected: localhost:6379
[order worker] 啟動，等待訂單訊息...
PublishOrder 成功: alice
PublishOrder 成功: bob
PublishOrder 成功: carol
PublishOrder 成功: dave
PublishOrder 成功: eve
等待 worker 寫入 DB...
完成，去 DB 確認 orders table 有沒有 5 筆資料
```

### 步驟二：確認資料寫入

```bash
docker exec -it etcd-ticket-postgres psql -U postgres -d etcd_ticket -c 'SELECT * FROM orders;'
```

應看到 5 筆訂單紀錄。
