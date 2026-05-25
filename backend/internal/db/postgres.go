package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	once sync.Once
	pool *pgxpool.Pool
)

// Init 在 main.go 啟動時呼叫一次，建立連線池
func Init(ctx context.Context, dsn string) error {
	var initErr error
	once.Do(func() {
		p, err := pgxpool.New(ctx, dsn)
		if err != nil {
			initErr = fmt.Errorf("postgres 連線失敗: %w", err)
			return
		}
		if err := p.Ping(ctx); err != nil {
			initErr = fmt.Errorf("postgres ping 失敗: %w", err)
			return
		}
		pool = p
		fmt.Println("postgres connected")
	})
	return initErr
}

// Get 供其他套件取得連線池
func Get() *pgxpool.Pool {
	if pool == nil {
		panic("postgres 尚未初始化，請先呼叫 db.Init()")
	}
	return pool
}

// Close 在程式結束時釋放連線池，於 main.go 呼叫
func Close() {
	if pool != nil {
		pool.Close()
	}
}
