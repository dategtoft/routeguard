package cache_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourusername/routeguard/cache"
)

func newCountingHandler(count *atomic.Int32, body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count.Add(1)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body)) //nolint:errcheck
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := cache.DefaultOptions()
	if opts.TTL != 60*time.Second {
		t.Errorf("expected TTL 60s, got %v", opts.TTL)
	}
	if len(opts.Methods) == 0 {
		t.Error("expected default methods to be non-empty")
	}
}

func TestNew_CachesGETResponse(t *testing.T) {
	var count atomic.Int32
	handler := cache.New(cache.DefaultOptions())(newCountingHandler(&count, "hello"))

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	}
	if count.Load() != 1 {
		t.Errorf("expected handler called once, got %d", count.Load())
	}
}

func TestNew_CacheHitHeader(t *testing.T) {
	var count atomic.Int32
	handler := cache.New(cache.DefaultOptions())(newCountingHandler(&count, "world"))

	req1 := httptest.NewRequest(http.MethodGet, "/hit", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Header().Get("X-Cache") != "MISS" {
		t.Errorf("first request should be MISS, got %s", rec1.Header().Get("X-Cache"))
	}

	req2 := httptest.NewRequest(http.MethodGet, "/hit", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Header().Get("X-Cache") != "HIT" {
		t.Errorf("second request should be HIT, got %s", rec2.Header().Get("X-Cache"))
	}
}

func TestNew_DoesNotCachePOST(t *testing.T) {
	var count atomic.Int32
	handler := cache.New(cache.DefaultOptions())(newCountingHandler(&count, "post"))

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/post", nil)
		handler.ServeHTTP(rec, req)
	}
	if count.Load() != 3 {
		t.Errorf("expected handler called 3 times for POST, got %d", count.Load())
	}
}

func TestNew_TTLExpiry(t *testing.T) {
	var count atomic.Int32
	opts := cache.Options{TTL: 50 * time.Millisecond, Methods: []string{http.MethodGet}}
	handler := cache.New(opts)(newCountingHandler(&count, "ttl"))

	req := httptest.NewRequest(http.MethodGet, "/ttl", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	time.Sleep(80 * time.Millisecond)

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/ttl", nil))
	if count.Load() != 2 {
		t.Errorf("expected handler called twice after TTL expiry, got %d", count.Load())
	}
}
