package bodysize_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/yourusername/routeguard/bodysize"
)

// ExampleNew demonstrates using bodysize middleware with default options.
func ExampleNew() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := bodysize.New(bodysize.DefaultOptions())
	h := mw(handler)

	req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader("small body"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	// Output: 200
}

// ExampleNew_customLimit demonstrates enforcing a 512-byte body limit.
func ExampleNew_customLimit() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	opts := bodysize.Options{
		MaxBytes:     512,
		ErrorMessage: "Payload too large, maximum 512 bytes allowed",
	}
	mw := bodysize.New(opts)
	h := mw(handler)

	// Simulate a request whose declared Content-Length exceeds the limit.
	req := httptest.NewRequest(http.MethodPost, "/data", strings.NewReader("x"))
	req.ContentLength = 1024
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	// Output: 413
}

// ExampleNew_getRequestSkipped demonstrates that GET requests without a body
// are passed through regardless of any configured limit.
func ExampleNew_getRequestSkipped() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	opts := bodysize.Options{
		MaxBytes:     1, // extremely small limit
		ErrorMessage: "body too large",
	}
	mw := bodysize.New(opts)
	h := mw(handler)

	// GET requests typically have no body, so the limit should not apply.
	req := httptest.NewRequest(http.MethodGet, "/resource", http.NoBody)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	// Output: 200
}
