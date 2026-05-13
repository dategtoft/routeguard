package observe_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/patrickward/routeguard/observe"
)

// ExampleNew demonstrates wrapping a handler with observe middleware
// using the default options. The /metrics path returns a JSON snapshot.
func ExampleNew() {
	app := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	h := observe.New(app, observe.DefaultOptions())

	// Simulate a normal request.
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/hello", nil))

	// Fetch the metrics snapshot.
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	fmt.Println(rec.Code)
	// Output:
	// 200
}

// ExampleNew_namespace shows how to prefix all metric keys with a namespace.
func ExampleNew_namespace() {
	app := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	opts := observe.Options{
		Path:      "/_internal/metrics",
		Namespace: "myapp",
	}
	h := observe.New(app, opts)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/_internal/metrics", nil))

	fmt.Println(rec.Code)
	// Output:
	// 200
}
