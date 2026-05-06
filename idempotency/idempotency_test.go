package idempotency_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/patrickward/routeguard/idempotency"
)

func newCountingHandler(t *testing.T) (http.Handler, *int64) {
	t.Helper()
	var count int64
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&count, 1)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created")) //nolint:errcheck
	}), &count
}

func TestDefaultOptions(t *testing.T) {
	opts := idempotency.DefaultOptions()
	if opts.Header != "Idempotency-Key" {
		t.Errorf("expected header Idempotency-Key, got %s", opts.Header)
	}
	if opts.TTL != 24*time.Hour {
		t.Errorf("expected TTL 24h, got %v", opts.TTL)
	}
	if len(opts.Methods) == 0 {
		t.Error("expected default methods to be non-empty")
	}
}

func TestNew_NoKey_AlwaysExecutes(t *testing.T) {
	handler, count := newCountingHandler(t)
	mw := idempotency.New(idempotency.DefaultOptions())(handler)

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		mw.ServeHTTP(rec, req)
	}

	if *count != 3 {
		t.Errorf("expected handler called 3 times, got %d", *count)
	}
}

func TestNew_SameKey_CachesResponse(t *testing.T) {
	handler, count := newCountingHandler(t)
	mw := idempotency.New(idempotency.DefaultOptions())(handler)

	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Idempotency-Key", "key-abc")
		mw.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", rec.Code)
		}
	}

	if *count != 1 {
		t.Errorf("expected handler called once, got %d", *count)
	}
}

func TestNew_ReplayedHeader(t *testing.T) {
	handler, _ := newCountingHandler(t)
	mw := idempotency.New(idempotency.DefaultOptions())(handler)

	sendRequest := func() *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Idempotency-Key", "replay-key")
		mw.ServeHTTP(rec, req)
		return rec
	}

	first := sendRequest()
	if first.Header().Get("X-Idempotency-Replayed") != "" {
		t.Error("first response should not have replayed header")
	}

	second := sendRequest()
	if second.Header().Get("X-Idempotency-Replayed") != "true" {
		t.Error("second response should have X-Idempotency-Replayed: true")
	}
}

func TestNew_DifferentKeys_IndependentCaches(t *testing.T) {
	handler, count := newCountingHandler(t)
	mw := idempotency.New(idempotency.DefaultOptions())(handler)

	for _, key := range []string{"key-1", "key-2", "key-3"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Idempotency-Key", key)
		mw.ServeHTTP(rec, req)
	}

	if *count != 3 {
		t.Errorf("expected 3 handler calls for 3 distinct keys, got %d", *count)
	}
}

func TestNew_NonCoveredMethod_NotCached(t *testing.T) {
	handler, count := newCountingHandler(t)
	mw := idempotency.New(idempotency.DefaultOptions())(handler)

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Idempotency-Key", "get-key")
		mw.ServeHTTP(rec, req)
	}

	if *count != 3 {
		t.Errorf("GET requests should not be cached, expected 3 calls, got %d", *count)
	}
}

func TestNew_TTLExpiry_ReExecutes(t *testing.T) {
	handler, count := newCountingHandler(t)
	opts := idempotency.DefaultOptions()
	opts.TTL = 50 * time.Millisecond
	mw := idempotency.New(opts)(handler)

	send := func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Idempotency-Key", "ttl-key")
		mw.ServeHTTP(rec, req)
	}

	send()
	time.Sleep(100 * time.Millisecond)
	send()

	if *count != 2 {
		t.Errorf("expected 2 handler calls after TTL expiry, got %d", *count)
	}
}
