package dedupe_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/patrickward/routeguard/dedupe"
)

// ExampleNew demonstrates the default deduplication middleware.
func ExampleNew() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "expensive response")
	})

	mw := dedupe.New(dedupe.DefaultOptions())(handler)

	req := httptest.NewRequest(http.MethodGet, "/data", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	// Output:
	// 200
}

// ExampleNew_customKeyFunc demonstrates using a custom key function
// so that requests with the same query parameter are coalesced regardless
// of other query string differences.
func ExampleNew_customKeyFunc() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "ok")
	})

	opts := dedupe.Options{
		KeyFunc: func(r *http.Request) string {
			// Coalesce by path only, ignoring query parameters.
			return r.Method + ":" + r.URL.Path
		},
	}

	mw := dedupe.New(opts)(handler)

	req := httptest.NewRequest(http.MethodGet, "/items?page=1", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	// Output:
	// 200
}
