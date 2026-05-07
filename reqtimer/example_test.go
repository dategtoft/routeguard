package reqtimer_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/joeydtaylor/routeguard/reqtimer"
)

// ExampleNew demonstrates attaching the request timer middleware with
// default options to a handler.
func ExampleNew() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := reqtimer.New(reqtimer.DefaultOptions())
	h := mw(handler)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	// Output:
	// 200
}

// ExampleNew_customOptions shows using microsecond precision and a custom header.
func ExampleNew_customOptions() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	mw := reqtimer.New(reqtimer.Options{
		Header:    "X-Elapsed",
		Precision: time.Microsecond,
	})
	h := mw(handler)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	h.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	// Output:
	// 200
}
