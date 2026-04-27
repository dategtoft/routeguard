package cache_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/yourusername/routeguard/cache"
)

// ExampleNew demonstrates using the cache middleware with default options.
func ExampleNew() {
	calls := 0
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "response")
	})

	handler := cache.New(cache.DefaultOptions())(backend)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	fmt.Printf("backend called %d time(s)\n", calls)
	// Output: backend called 1 time(s)
}

// ExampleNew_customTTL demonstrates configuring a short TTL.
func ExampleNew_customTTL() {
	opts := cache.Options{
		TTL:     500 * time.Millisecond,
		Methods: []string{http.MethodGet},
	}

	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := cache.New(opts)(backend)
	req := httptest.NewRequest(http.MethodGet, "/short", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	fmt.Println(rec.Header().Get("X-Cache"))
	// Output: MISS
}
