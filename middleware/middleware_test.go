package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/routeguard/jwt"
	"github.com/yourusername/routeguard/middleware"
	"github.com/yourusername/routeguard/ratelimit"
)

const testSecret = "test-secret-key"

// newTestHandler returns a simple HTTP handler that responds with 200 OK.
func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// TestNew_PassThrough verifies that a middleware chain with no options
// passes requests through to the underlying handler unchanged.
func TestNew_PassThrough(t *testing.T) {
	m := middleware.New()
	handler := m.Handler(newTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

// TestNew_WithRateLimit verifies that the combined middleware enforces
// the rate limit when the RateLimit option is provided.
func TestNew_WithRateLimit(t *testing.T) {
	rl := ratelimit.New(2, time.Second)
	m := middleware.New(middleware.WithRateLimit(rl))
	handler := m.Handler(newTestHandler())

	// First two requests should succeed.
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	// Third request should be rate-limited.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 after limit exceeded, got %d", rec.Code)
	}
}

// TestNew_WithJWT verifies that the combined middleware rejects requests
// without a valid JWT when the JWT option is provided.
func TestNew_WithJWT(t *testing.T) {
	j := jwt.New(testSecret)
	m := middleware.New(middleware.WithJWT(j))
	handler := m.Handler(newTestHandler())

	// Request without token should be rejected.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without token, got %d", rec.Code)
	}
}

// TestNew_WithJWT_ValidToken verifies that a valid JWT token allows
// the request to pass through.
func TestNew_WithJWT_ValidToken(t *testing.T) {
	j := jwt.New(testSecret)
	tokenStr, err := j.Generate("user-42", time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	m := middleware.New(middleware.WithJWT(j))
	handler := m.Handler(newTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 with valid token, got %d", rec.Code)
	}
}

// TestNew_WithBoth verifies that both rate limiting and JWT validation
// are applied when both options are provided.
func TestNew_WithBoth(t *testing.T) {
	j := jwt.New(testSecret)
	rl := ratelimit.New(5, time.Second)
	m := middleware.New(middleware.WithJWT(j), middleware.WithRateLimit(rl))
	handler := m.Handler(newTestHandler())

	// No token — should fail JWT check first.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:9999"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without token, got %d", rec.Code)
	}

	// Valid token — should succeed.
	tokenStr, _ := j.Generate("user-1", time.Hour)
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "10.0.0.1:9999"
	req2.Header.Set("Authorization", "Bearer "+tokenStr)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("expected 200 with valid token, got %d", rec2.Code)
	}
}
