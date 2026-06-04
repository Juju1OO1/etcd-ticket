# internal/repository

資料庫操作封裝層，負責執行 SQL，讓 service 層不需要知道 DB 細節。

---

## order_repo.go

### InsertOrder(ctx, order)

把一筆 `model.Order` 寫進 PostgreSQL 的 `orders` 表。

使用 `$1, $2...` 參數佔位符，由 pgx 處理 SQL injection 防護。
`id` 和 `created_at` 不需要傳入，由 DB 自動產生。

若同一個 `user_name + area` 已有紀錄（UNIQUE constraint），使用 `ON CONFLICT DO NOTHING` 靜默略過，不報錯。

呼叫方：`service/order_service.go` 的 `processMessage()`。

### GetOrdersByUser(ctx, userName)

查詢指定用戶的所有購票紀錄，依 `created_at` 降冪排列（最新在前）。

回傳 `[]model.Order`，若無任何紀錄回傳空 slice（`[]`），不回傳 nil，確保 JSON 序列化為 `[]` 而非 `null`。

呼叫方：`api/order.go` 的 `OrdersHandler()`。
