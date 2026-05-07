package requestsize_test

import (
	"fmt"
	"net/http"

	"github.com/joeychilson/routeguard/requestsize"
)

// ExampleNew demonstrates attaching the middleware with default options.
func ExampleNew() {
	// Reject any POST/PUT/PATCH request whose body exceeds 1 MB (the default).
	mw := requestsize.New(requestsize.DefaultOptions())

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "request accepted")
	}))

	http.Handle("/upload", handler)
}

// ExampleNew_customLimit shows how to set a 64 KB limit and allow only POST.
func ExampleNew_customLimit() {
	opts := requestsize.Options{
		MaxBytes:     64 * 1024, // 64 KB
		ErrorMessage: "payload too large",
		SkipMethods:  []string{http.MethodGet, http.MethodHead},
	}

	mw := requestsize.New(opts)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "small request accepted")
	}))

	http.Handle("/api/data", handler)
}
