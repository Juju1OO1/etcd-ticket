# internal/db

PostgreSQL 連線管理與資料庫初始化設定。

---

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

| 欄位 | 型別 | 說明 |
|------|------|------|
| id | BIGSERIAL | 自動遞增主鍵 |
| order_id | VARCHAR(36) | UUID，唯一訂單識別碼 |
| user_name | VARCHAR(100) | 購票用戶名稱 |
| phone_num | VARCHAR(20) | 購票用戶電話 |
| area | INT | 購票區域編號 |
| status | VARCHAR(20) | 訂單狀態（success / cancelled），預設 success |
| created_at | TIMESTAMPTZ | 建立時間，DB 自動填入 |

索引：
- `idx_orders_user` — 依 user_name 查詢
- `idx_orders_area` — 依 area 查詢

---

## docker-compose.yaml

一鍵啟動 PostgreSQL 和 Redis 容器。

| 服務 | Image | Port |
|------|-------|------|
| postgres | postgres:16 | 5433（對外）→ 5432（容器內） |
| redis | redis:7 | 6379 |

PostgreSQL 使用 5433 對外，避免與本機預設 5432 衝突。
schema.sql 掛載進 `/docker-entrypoint-initdb.d/`，第一次啟動自動建表。

### 啟動指令
```bash
cd backend/internal/db
docker compose up -d
```
