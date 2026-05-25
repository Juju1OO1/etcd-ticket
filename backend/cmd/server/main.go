package main

import (
	"context"
	"etcd-ticket/internal/api"
	"etcd-ticket/internal/db"
	"etcd-ticket/internal/etcd"
	"etcd-ticket/internal/mq"
	"etcd-ticket/internal/service"
	"etcd-ticket/internal/watcher"
	"fmt"
	"os"

	"sigs.k8s.io/yaml"
)

type AppConfig struct {
	Server struct {
		Port int `json:"port"`
	} `json:"server"`
	Database struct {
		DSN string `yaml:"dsn"`
	} `yaml:"database"`
	Redis struct {
		Addr string `yaml:"addr"`
	} `yaml:"redis"`
	RateLimit struct {
		RPS   float64 `json:"rps"`
		Burst int     `json:"burst"`
	} `json:"rate_limit"`
}

func loadAppConfig(path string) (AppConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return AppConfig{}, err
	}
	var cfg AppConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return AppConfig{}, err
	}
	return cfg, nil
}

func main() {
	ctx := context.Background()

	cfg, err := loadAppConfig("./configs/config.yaml")
	if err != nil {
		panic(fmt.Sprintf("讀取 config 失敗: %v", err))
	}

	// 初始化 PostgreSQL
	if err := db.Init(ctx, cfg.Database.DSN); err != nil {
		panic(err)
	}
	defer db.Close()

	// 初始化 Redis Stream
	if err := mq.Init(ctx, cfg.Redis.Addr); err != nil {
		panic(err)
	}
	defer mq.Close()

	// 啟動訂單非同步 worker
	service.StartOrderWorker(ctx)

	// 初始化 etcd 售票資料
	areas := []service.AreaConfig{
		{Name: "area1", Limit: 100},
		{Name: "area2", Limit: 50},
	}
	if err := service.InitEtcd(ctx, areas); err != nil {
		panic(fmt.Sprintf("etcd 初始化失敗: %v", err))
	}
	defer etcd.Close()

	errCh, err := watcher.StartHTTPClient(ctx, []int{1, 2}, "127.0.0.1", 8888)
	if err != nil {
		panic(fmt.Sprintf("啟動剩餘票數 HTTP client 失敗: %v", err))
	}

	go func() {
		for err := range errCh {
			fmt.Println("送剩餘票數到 websocket server 失敗:", err)
		}
	}()

	// 啟動 HTTP server（API Gateway）
	router := api.NewRouter(api.RateLimitConfig{
		RPS:   cfg.RateLimit.RPS,
		Burst: cfg.RateLimit.Burst,
	})
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	fmt.Printf("Rate limit 設定：rps=%.0f burst=%d\n", cfg.RateLimit.RPS, cfg.RateLimit.Burst)
	fmt.Printf("系統啟動完成，HTTP server 監聽 %s\n", addr)
	if err := router.Run(addr); err != nil {
		panic(fmt.Sprintf("HTTP server 啟動失敗: %v", err))
	}
}
