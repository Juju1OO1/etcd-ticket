package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

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

func TestHealthz_200(t *testing.T) {
	r := NewRouter(RateLimitConfig{RPS: 100, Burst: 100})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, newReq(http.MethodGet, "/healthz"))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestReserveStub_501(t *testing.T) {
	r := NewRouter(RateLimitConfig{RPS: 100, Burst: 100})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, newReq(http.MethodPost, "/api/tickets/reserve"))
	if w.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestCheckoutStub_501(t *testing.T) {
	r := NewRouter(RateLimitConfig{RPS: 100, Burst: 100})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, newReq(http.MethodPost, "/api/tickets/checkout"))
	if w.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d body=%s", w.Code, w.Body.String())
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
	r := NewRouter(RateLimitConfig{RPS: 1, Burst: 5})

	var pass, blocked int
	for i := 0; i < 50; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, newReq(http.MethodGet, "/api/tickets/status"))
		switch w.Code {
		case http.StatusTooManyRequests:
			blocked++
		case http.StatusNotImplemented: // stub
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
