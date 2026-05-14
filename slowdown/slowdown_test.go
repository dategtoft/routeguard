package slowdown_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jcosta33/routeguard/slowdown"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := slowdown.DefaultOptions()
	if opts.Delay != 500*time.Millisecond {
		t.Errorf("expected 500ms delay, got %v", opts.Delay)
	}
	if len(opts.OnlyPaths) != 0 {
		t.Errorf("expected no OnlyPaths, got %v", opts.OnlyPaths)
	}
}

func TestNew_DelaysRequest(t *testing.T) {
	opts := slowdown.Options{Delay: 50 * time.Millisecond}
	mw := slowdown.New(opts)(newTestHandler())

	start := time.Now()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)
	elapsed := time.Since(start)

	if elapsed < 50*time.Millisecond {
		t.Errorf("expected at least 50ms delay, got %v", elapsed)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestNew_OnlyPaths_MatchingPath_Delayed(t *testing.T) {
	opts := slowdown.Options{
		Delay:     50 * time.Millisecond,
		OnlyPaths: []string{"/slow"},
	}
	mw := slowdown.New(opts)(newTestHandler())

	start := time.Now()
	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)
	elapsed := time.Since(start)

	if elapsed < 50*time.Millisecond {
		t.Errorf("expected delay on /slow, got %v", elapsed)
	}
}

func TestNew_OnlyPaths_NonMatchingPath_NotDelayed(t *testing.T) {
	opts := slowdown.Options{
		Delay:     200 * time.Millisecond,
		OnlyPaths: []string{"/slow"},
	}
	mw := slowdown.New(opts)(newTestHandler())

	start := time.Now()
	req := httptest.NewRequest(http.MethodGet, "/fast", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)
	elapsed := time.Since(start)

	if elapsed >= 200*time.Millisecond {
		t.Errorf("expected no delay on /fast, got %v", elapsed)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestNew_CancelledContext_Returns503(t *testing.T) {
	opts := slowdown.Options{Delay: 500 * time.Millisecond}
	mw := slowdown.New(opts)(newTestHandler())

	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 on cancelled context, got %d", rec.Code)
	}
}
