package idempotency_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/patrickward/routeguard/idempotency"
)

// ExampleNew demonstrates idempotency middleware with default options.
// A POST request carrying the same Idempotency-Key header will only
// execute the underlying handler once; subsequent requests receive
// the cached response.
func ExampleNew() {
	var calls int
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintln(w, "order created")
	})

	mw := idempotency.New(idempotency.DefaultOptions())(handler)

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/orders", nil)
		req.Header.Set("Idempotency-Key", "order-xyz-123")
		mw.ServeHTTP(rec, req)
	}

	fmt.Printf("handler executed %d time(s)\n", calls)
	// Output:
	// handler executed 1 time(s)
}

// ExampleNew_customOptions shows how to configure a short TTL and restrict
// idempotency checks to PATCH requests only.
func ExampleNew_customOptions() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "patched")
	})

	opts := idempotency.Options{
		Header:  "X-Request-Key",
		TTL:     10 * time.Minute,
		Methods: []string{http.MethodPatch},
	}

	mw := idempotency.New(opts)(handler)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/resource/1", nil)
	req.Header.Set("X-Request-Key", "patch-unique-key")
	mw.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	// Output:
	// 200
}
