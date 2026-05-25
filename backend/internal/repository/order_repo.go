// 把一筆 Order 寫進 DB
package repository

import (
	"context"
	"etcd-ticket/internal/db"
	"etcd-ticket/internal/model"
)

func InsertOrder(ctx context.Context, order model.Order) error {
	sql := `
		INSERT INTO orders (order_id, user_name, phone_num, area, status)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_name, area) DO NOTHING
	`
	_, err := db.Get().Exec(ctx, sql,
		order.OrderID,
		order.UserName,
		order.PhoneNum,
		order.Area,
		order.Status,
	)
	return err
}
