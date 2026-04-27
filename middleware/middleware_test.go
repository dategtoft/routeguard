package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/routeguard/cors"
	"github.com/yourusername/routeguard/jwt"
	"github.com/yourusername/routeguard/logger"
	"github.com/yourusername/routeguard/middleware"
	"github.com/yourusername/routeguard/ratelimit"
	"github.com/yourusername/routeguard/recovery"
	"github.com/yourusername/routeguard/timeout"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok")) //nolint:errcheck
	})
}

func TestNew_PassThrough(t *testing.T) {
	mw := middleware.New(middleware.Options{})
	handler := mw(newTestHandler())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestNew_WithRateLimit(t *testing.T) {
	rlOpts := ratelimit.Options{Limit: 1, Window: time.Minute}
	mw := middleware.New(middleware.Options{RateLimit: &rlOpts})
	handler := mw(newTestHandler())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("first request: expected 200, got %d", rr.Code)
	}

	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req)
	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("second request: expected 429, got %d", rr2.Code)
	}
}

func TestNew_WithJWT(t *testing.T) {
	jwtOpts := jwt.Options{Secret: "test-secret"}
	mw := middleware.New(middleware.Options{JWT: &jwtOpts})
	handler := mw(newTestHandler())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without token, got %d", rr.Code)
	}
}

func TestNew_WithTimeout(t *testing.T) {
	timeoutOpts := timeout.Options{
		Duration:   50 * time.Millisecond,
		Message:    "timed out",
		StatusCode: http.StatusGatewayTimeout,
	}
	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(200 * time.Millisecond):
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
		}
	})
	mw := middleware.New(middleware.Options{Timeout: &timeoutOpts})
	handler := mw(slowHandler)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusGatewayTimeout {
		t.Errorf("expected 504, got %d", rr.Code)
	}
}

func TestNew_WithAllOptions(t *testing.T) {
	rlOpts := ratelimit.Options{Limit: 10, Window: time.Minute}
	logOpts := logger.Options{Prefix: "[test]"}
	corsOpts := cors.DefaultOptions()
	recOpts := recovery.DefaultOptions()
	timeoutOpts := timeout.Options{Duration: 5 * time.Second}

	mw := middleware.New(middleware.Options{
		RateLimit: &rlOpts,
		Logger:    &logOpts,
		CORS:      &corsOpts,
		Recovery:  &recOpts,
		Timeout:   &timeoutOpts,
	})
	handler := mw(newTestHandler())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}
