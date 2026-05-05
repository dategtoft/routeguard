package proxy_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/patrickward/routeguard/proxy"
)

func ExampleNew() {
	// Simulate a backend service.
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "backend response")
	}))
	defer backend.Close()

	// Create a proxy middleware that forwards all requests to the backend.
	mw := proxy.New(proxy.DefaultOptions(backend.URL))

	// Wrap your router or handler.
	handler := mw(http.NotFoundHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	fmt.Println(rec.Body.String())
	// Output: backend response
}

func ExampleNew_stripPrefix() {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.URL.Path)
	}))
	defer backend.Close()

	opts := proxy.DefaultOptions(backend.URL)
	opts.StripPrefix = "/api"

	mw := proxy.New(opts)
	handler := mw(http.NotFoundHandler())

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	fmt.Println(rec.Body.String())
	// Output: /users
}
