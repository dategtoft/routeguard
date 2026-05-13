package nocache_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/patrickward/routeguard/nocache"
)

func ExampleNew() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello")
	})

	// Wrap with default nocache options.
	mw := nocache.New(nocache.DefaultOptions())(handler)

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(rec.Header().Get("Cache-Control"))
	fmt.Println(rec.Header().Get("Pragma"))
	fmt.Println(rec.Header().Get("Expires"))
	// Output:
	// no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0
	// no-cache
	// 0
}

func ExampleNew_noPragma() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello")
	})

	opts := nocache.DefaultOptions()
	opts.Pragma = false

	mw := nocache.New(opts)(handler)

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(rec.Header().Get("Pragma") == "")
	// Output:
	// true
}
