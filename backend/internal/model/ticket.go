// 定義一筆訂單長什麼樣子
package model

import "time"

type Order struct {
	ID        int64     `db:"id"`
	OrderID   string    `db:"order_id"`  // UUID
	UserName  string    `db:"user_name"`
	PhoneNum  string    `db:"phone_num"`
	Area      int       `db:"area"`
	Status    string    `db:"status"`    // "success" | "cancelled"
	CreatedAt time.Time `db:"created_at"`
}
