package retry_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/patrickward/routeguard/retry"
)

func ExampleNew() {
	// Wrap a handler with default retry behaviour (3 attempts, 100 ms delay).
	handler := retry.New(retry.DefaultOptions())(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	fmt.Println(rec.Code)
	// Output: 200
}

func ExampleNew_customOptions() {
	// Retry up to 5 times with a 50 ms delay, only on 503.
	opts := retry.Options{
		MaxAttempts: 5,
		Delay:       50 * time.Millisecond,
		ShouldRetry: func(code int) bool {
			return code == http.StatusServiceUnavailable
		},
	}

	handler := retry.New(opts)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	fmt.Println(rec.Code)
	// Output: 200
}
