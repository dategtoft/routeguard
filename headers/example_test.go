package headers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/yourusername/routeguard/headers"
)

func ExampleNew() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Use default security headers.
	mw := headers.New(headers.DefaultOptions())

	rec := httptest.NewRecorder()
	mw(handler).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(rec.Header().Get("X-Frame-Options"))
	// Output: DENY
}

func ExampleNew_customHeaders() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Add custom headers alongside security headers.
	mw := headers.New(headers.Options{
		SecurityHeaders: true,
		HSTSMaxAge:      31536000,
		Custom: map[string]string{
			"X-App-Version": "2.0.0",
		},
	})

	rec := httptest.NewRecorder()
	mw(handler).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(rec.Header().Get("X-App-Version"))
	// Output: 2.0.0
}
