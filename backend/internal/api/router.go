package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// RateLimitConfig 由 main.go 從 config.yaml 讀出後傳入。
type RateLimitConfig struct {
	RPS   float64
	Burst int
}

// NewRouter 組裝整個 HTTP 路由：
//   - middleware：Logger / Recovery / CORS / RateLimit
//   - GET /healthz（探活，不掛 rate limit）
//   - /api/tickets/{reserve,checkout,status}（成員 4 填 body）

func NewRouter(rl RateLimitConfig) *gin.Engine {
	// 除錯
	fmt.Printf(
		"RateLimit Config => RPS=%v Burst=%v\n",
		rl.RPS,
		rl.Burst,
	)
	// 除錯

	r := gin.New()
	// 沒有真正的反向代理，明確拒絕信任，避免 X-Forwarded-For 被偽造繞過 rate limit。
	_ = r.SetTrustedProxies(nil)

	limiter := NewIPRateLimiter(rl.RPS, rl.Burst)

	r.Use(gin.Logger(), gin.Recovery(), CORSMiddleware())

	// 探活不掛 rate limit，避免 k8s / 監控誤判
	r.GET("/healthz", HealthzHandler(limiter))

	apiGroup := r.Group("/api", RateLimitMiddleware(limiter))
	{
		tickets := apiGroup.Group("/tickets")
		{
			tickets.POST("/reserve", ReserveHandler)
			tickets.POST("/checkout", CheckoutHandler)
			tickets.GET("/status", StatusHandler)
		}
	}

	return r
}
