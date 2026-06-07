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

func GetOrdersByUser(ctx context.Context, userName string) ([]model.Order, error) {
	sql := `
		SELECT id, order_id, user_name, phone_num, area, status, created_at
		FROM orders
		WHERE user_name = $1
		ORDER BY created_at DESC
	`
	rows, err := db.Get().Query(ctx, sql, userName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]model.Order, 0)
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(
			&o.ID,
			&o.OrderID,
			&o.UserName,
			&o.PhoneNum,
			&o.Area,
			&o.Status,
			&o.CreatedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}
