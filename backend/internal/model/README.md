# internal/model

資料結構定義，是 Go 程式和 PostgreSQL 之間的共同語言。

---

## ticket.go

### Order struct
對應 PostgreSQL `orders` 表的一筆訂單紀錄。

| 欄位 | 型別 | db tag | 說明 |
|------|------|--------|------|
| ID | int64 | id | 自動遞增主鍵（DB 產生，不需要傳入）|
| OrderID | string | order_id | UUID 唯一識別碼 |
| UserName | string | user_name | 購票用戶名稱 |
| PhoneNum | string | phone_num | 購票用戶電話 |
| Area | int | area | 購票區域編號 |
| Status | string | status | 訂單狀態（success / cancelled）|
| CreatedAt | time.Time | created_at | 建立時間（DB 產生，不需要傳入）|
