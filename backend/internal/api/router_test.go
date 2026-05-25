package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"etcd-ticket/internal/service"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func newReq(method, path string) *http.Request {
	r := httptest.NewRequest(method, path, nil)
	// 強制覆寫 RemoteAddr，確保所有測試共用同一 IP（觸發 per-IP rate limit）
	r.RemoteAddr = "10.0.0.1:1234"
	return r
}

func newJSONReq(method, path, body string) *http.Request {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.RemoteAddr = "10.0.0.1:1234"
	return r
}

func stubHandlers(t *testing.T) {
	t.Helper()

	prevReserve := reserveTicketFn
	prevCheckout := checkoutTicketFn
	prevPublish := publishOrderFn
	prevAvailable := availableTicketsFn
	prevLoadStatus := loadAreaStatusFn

	reserveTicketFn = func(context.Context, service.TicketData) (bool, error) {
		return true, nil
	}
	checkoutTicketFn = func(context.Context, service.TicketData) error {
		return nil
	}
	publishOrderFn = func(context.Context, service.TicketData) error {
		return nil
	}
	availableTicketsFn = func(context.Context, int) (int64, error) {
		return 42, nil
	}
	loadAreaStatusFn = func(context.Context, int) (string, error) {
		return "on", nil
	}

	t.Cleanup(func() {
		reserveTicketFn = prevReserve
		checkoutTicketFn = prevCheckout
		publishOrderFn = prevPublish
		availableTicketsFn = prevAvailable
		loadAreaStatusFn = prevLoadStatus
	})
}

func TestHealthz_200(t *testing.T) {
	r := NewRouter(RateLimitConfig{RPS: 100, Burst: 100})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, newReq(http.MethodGet, "/healthz"))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestReserveOK(t *testing.T) {
	stubHandlers(t)
	r := NewRouter(RateLimitConfig{RPS: 100, Burst: 100})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, newJSONReq(http.MethodPost, "/api/tickets/reserve", `{"user_name":"alice","phone_num":"0911111111","area":1}`))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestCheckoutOK(t *testing.T) {
	stubHandlers(t)
	r := NewRouter(RateLimitConfig{RPS: 100, Burst: 100})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, newJSONReq(http.MethodPost, "/api/tickets/checkout", `{"user_name":"alice","phone_num":"0911111111","area":1}`))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestStatusOK(t *testing.T) {
	stubHandlers(t)
	r := NewRouter(RateLimitConfig{RPS: 100, Burst: 100})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, newReq(http.MethodGet, "/api/tickets/status"))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestUnknownPath_404(t *testing.T) {
	r := NewRouter(RateLimitConfig{RPS: 100, Burst: 100})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, newReq(http.MethodGet, "/nope"))
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// 確認 rate limit middleware 真的會擋：burst=5、rps=1，連發 50 次。
// 因為 burst=5，前 5 個可通過；rps=1 表示一秒只補 1 token，迴圈在 ms 內完成，
// 所以幾乎所有後續請求都會被擋。
func TestRateLimit_Blocks(t *testing.T) {
	stubHandlers(t)
	r := NewRouter(RateLimitConfig{RPS: 1, Burst: 5})

	var pass, blocked int
	for i := 0; i < 50; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, newReq(http.MethodGet, "/api/tickets/status"))
		switch w.Code {
		case http.StatusTooManyRequests:
			blocked++
		case http.StatusOK:
			pass++
		default:
			t.Fatalf("unexpected status %d body=%s", w.Code, w.Body.String())
		}
	}

	if pass < 5 || pass > 6 {
		t.Fatalf("expected ~5 pass (burst=5), got %d", pass)
	}
	if blocked < 40 {
		t.Fatalf("expected >=40 blocked, got %d", blocked)
	}
}

// healthz 不掛 rate limit，連發 50 次應該全部通過
func TestHealthz_NotRateLimited(t *testing.T) {
	r := NewRouter(RateLimitConfig{RPS: 1, Burst: 1})

	for i := 0; i < 50; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, newReq(http.MethodGet, "/healthz"))
		if w.Code != http.StatusOK {
			t.Fatalf("healthz request %d failed with %d", i, w.Code)
		}
	}
}
