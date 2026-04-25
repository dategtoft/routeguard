package ratelimit_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/routeguard/ratelimit"
)

func TestAllow_WithinLimit(t *testing.T) {
	l := ratelimit.New(3, time.Minute)
	for i := 0; i < 3; i++ {
		if !l.Allow("client1") {
			t.Fatalf("expected request %d to be allowed", i+1)
		}
	}
}

func TestAllow_ExceedsLimit(t *testing.T) {
	l := ratelimit.New(2, time.Minute)
	l.Allow("client2")
	l.Allow("client2")
	if l.Allow("client2") {
		t.Fatal("expected request to be blocked after exceeding rate limit")
	}
}

func TestAllow_WindowReset(t *testing.T) {
	l := ratelimit.New(1, 50*time.Millisecond)
	if !l.Allow("client3") {
		t.Fatal("first request should be allowed")
	}
	if l.Allow("client3") {
		t.Fatal("second request should be blocked within window")
	}
	time.Sleep(60 * time.Millisecond)
	if !l.Allow("client3") {
		t.Fatal("request after window reset should be allowed")
	}
}

func TestMiddleware_Blocks(t *testing.T) {
	l := ratelimit.New(1, time.Minute)
	handler := l.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req)
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rr2.Code)
	}
}
