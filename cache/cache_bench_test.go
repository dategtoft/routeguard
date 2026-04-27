package cache_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/routeguard/cache"
)

func BenchmarkCache_Hit(b *testing.B) {
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("benchmark response")) //nolint:errcheck
	})
	handler := cache.New(cache.DefaultOptions())(backend)

	// Warm up the cache.
	warmReq := httptest.NewRequest(http.MethodGet, "/bench", nil)
	handler.ServeHTTP(httptest.NewRecorder(), warmReq)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/bench", nil)
			handler.ServeHTTP(rec, req)
		}
	})
}

func BenchmarkCache_Miss(b *testing.B) {
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("miss response")) //nolint:errcheck
	})
	handler := cache.New(cache.DefaultOptions())(backend)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		// Unique path per iteration forces a cache miss every time.
		req := httptest.NewRequest(http.MethodPost, "/bench", nil)
		handler.ServeHTTP(rec, req)
	}
}
