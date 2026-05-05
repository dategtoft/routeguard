package stripprefix_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/yourusername/routeguard/stripprefix"
)

// ExampleNew demonstrates stripping a version prefix from all API routes.
func ExampleNew() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, r.URL.Path)
	})

	mw := stripprefix.New(stripprefix.Options{
		Prefix: "/v1",
	})(handler)

	req := httptest.NewRequest(http.MethodGet, "/v1/users", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	fmt.Print(rec.Body.String())
	// Output:
	// /users
}

// ExampleNew_mismatch404 shows how to return 404 for unmatched prefixes.
func ExampleNew_mismatch404() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := stripprefix.New(stripprefix.Options{
		Prefix:             "/v1",
		RedirectOnMismatch: true,
	})(handler)

	req := httptest.NewRequest(http.MethodGet, "/v2/users", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	// Output:
	// 404
}
