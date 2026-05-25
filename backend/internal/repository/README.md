# internal/repository

資料庫操作封裝層，負責執行 SQL，讓 service 層不需要知道 DB 細節。

---

## order_repo.go

### InsertOrder(ctx, order)
把一筆 `model.Order` 寫進 PostgreSQL 的 `orders` 表。

使用 `$1, $2...` 參數佔位符，由 pgx 處理 SQL injection 防護。
`id` 和 `created_at` 不需要傳入，由 DB 自動產生。

呼叫方：`service/order_service.go` 的 `processMessage()`。
