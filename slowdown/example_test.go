package slowdown_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/jcosta33/routeguard/slowdown"
)

// ExampleNew demonstrates applying a uniform delay to all requests.
func ExampleNew() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := slowdown.New(slowdown.Options{
		Delay: 100 * time.Millisecond,
	})

	ts := httptest.NewServer(mw(handler))
	defer ts.Close()

	start := time.Now()
	resp, _ := http.Get(ts.URL + "/")
	elapsed := time.Since(start)

	fmt.Println(resp.StatusCode)
	fmt.Println(elapsed >= 100*time.Millisecond)
	// Output:
	// 200
	// true
}

// ExampleNew_specificPaths demonstrates delaying only selected paths.
func ExampleNew_specificPaths() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := slowdown.New(slowdown.Options{
		Delay:     200 * time.Millisecond,
		OnlyPaths: []string{"/api/slow"},
	})

	ts := httptest.NewServer(mw(handler))
	defer ts.Close()

	resp, _ := http.Get(ts.URL + "/api/fast")
	fmt.Println(resp.StatusCode)
	// Output:
	// 200
}
