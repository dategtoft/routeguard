package cloneheader_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/jaxron/routeguard/cloneheader"
)

// ExampleNew demonstrates copying a proxy-injected header to a canonical name.
func ExampleNew() {
	handler := cloneheader.New(
		cloneheader.Options{
			Rules: []cloneheader.Rule{
				{Source: "X-Forwarded-User", Destination: "X-User"},
			},
		},
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, r.Header.Get("X-User"))
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-User", "alice")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	fmt.Print(rec.Body.String())
	// Output: alice
}

// ExampleNew_overwrite demonstrates replacing an existing destination header.
func ExampleNew_overwrite() {
	handler := cloneheader.New(
		cloneheader.Options{
			Rules:     []cloneheader.Rule{{Source: "X-Real-Role", Destination: "X-Role"}},
			Overwrite: true,
		},
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, r.Header.Get("X-Role"))
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-Role", "admin")
	req.Header.Set("X-Role", "user")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	fmt.Print(rec.Body.String())
	// Output: admin
}
