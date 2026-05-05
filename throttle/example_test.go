package throttle_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/patrickward/routeguard/throttle"
)

func ExampleNew() {
	opts := throttle.DefaultOptions()
	opts.MaxConcurrent = 50
	opts.MaxQueueSize = 25
	opts.Timeout = 3 * time.Second

	mw := throttle.New(opts)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello")
	}))

	http.Handle("/", handler)
}

func ExampleNew_strict() {
	// Allow only 10 simultaneous requests with no queue; reject extras immediately.
	mw := throttle.New(throttle.Options{
		MaxConcurrent: 10,
		MaxQueueSize:  0,
		Timeout:       0,
		StatusCode:    http.StatusTooManyRequests,
		Message:       `{"error":"server busy"}`,
	})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	http.Handle("/api/", handler)
}
