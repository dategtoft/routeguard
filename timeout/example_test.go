package timeout_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/yourusername/routeguard/timeout"
)

// ExampleNew demonstrates basic usage of the timeout middleware.
func ExampleNew() {
	opts := timeout.Options{
		Duration:   100 * time.Millisecond,
		Message:    "request timed out",
		StatusCode: http.StatusGatewayTimeout,
	}

	// A handler that responds immediately.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "hello")
	})

	mw := timeout.New(opts)
	protected := mw(handler)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	protected.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String())
	// Output:
	// 200
	// hello
}
