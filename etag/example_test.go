package etag_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/yourusername/routeguard/etag"
)

func ExampleNew() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!")) //nolint:errcheck
	})

	mw := etag.New(etag.DefaultOptions())
	protected := mw(handler)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	protected.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Header().Get("ETag") != "")
	// Output:
	// 200
	// true
}

func ExampleNew_weakETag() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!")) //nolint:errcheck
	})

	mw := etag.New(etag.Options{Weak: true})
	protected := mw(handler)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	protected.ServeHTTP(rec, req)

	e := rec.Header().Get("ETag")
	fmt.Println(e[:2])
	// Output:
	// W/
}
