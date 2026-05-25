package api

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// visitor 代表單一 IP 的限流狀態。
type visitor struct {
	limiter  *rate.Limiter
	lastSeen atomic.Int64 // unix nano，避免每次更新都要 lock
}

// IPRateLimiter 是 per-IP token bucket 限流器。
// 用 sync.Map 存 ip -> *visitor；背景 goroutine 每 3 分鐘清掉超過 10 分鐘沒活動的 entry。
type IPRateLimiter struct {
	visitors sync.Map
	rps      rate.Limit
	burst    int
	blocked  atomic.Int64 // 累計被擋次數
}

// NewIPRateLimiter 建立並啟動清理 goroutine。
func NewIPRateLimiter(rps float64, burst int) *IPRateLimiter {
	l := &IPRateLimiter{
		rps:   rate.Limit(rps),
		burst: burst,
	}
	go l.cleanupLoop()
	return l
}

func (l *IPRateLimiter) get(ip string) *visitor {
	if v, ok := l.visitors.Load(ip); ok {
		nv := v.(*visitor)
		nv.lastSeen.Store(time.Now().UnixNano())
		return nv
	}
	nv := &visitor{limiter: rate.NewLimiter(l.rps, l.burst)}
	nv.lastSeen.Store(time.Now().UnixNano())
	actual, _ := l.visitors.LoadOrStore(ip, nv)
	return actual.(*visitor)
}

// Allow 回傳該 IP 是否仍有 token 可用。
func (l *IPRateLimiter) Allow(ip string) bool {
	return l.get(ip).limiter.Allow()
}

// Blocked 回傳累計被擋次數（用於 /healthz 觀察壓測結果）。
func (l *IPRateLimiter) Blocked() int64 {
	return l.blocked.Load()
}

// cleanupLoop 每 3 分鐘掃一次 sync.Map，移除 10 分鐘內無活動的 IP，避免長跑/壓測時記憶體洩漏。
func (l *IPRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-10 * time.Minute).UnixNano()
		l.visitors.Range(func(k, val interface{}) bool {
			if val.(*visitor).lastSeen.Load() < cutoff {
				l.visitors.Delete(k)
			}
			return true
		})
	}
}

// RateLimitMiddleware 是 per-IP 限流的 gin middleware。
func RateLimitMiddleware(l *IPRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !l.Allow(ip) {
			l.blocked.Add(1)
			c.Header("Retry-After", "1")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code": http.StatusTooManyRequests,
				"msg":  "too many requests",
			})
			return
		}
		c.Next()
	}
}

// CORSMiddleware 允許前端（Vite dev server）跨來源呼叫。
// 學期 project 場景，無敏感 cookie，直接放行所有 origin。
func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	})
}

// HealthzHandler 是 GET /healthz 探活端點。
// 順便回傳 rate limit 累計被擋次數，方便 demo / 壓測現場觀察。
func HealthzHandler(l *IPRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":        "ok",
			"blocked_total": l.Blocked(),
		})
	}
}
