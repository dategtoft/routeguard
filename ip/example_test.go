package ip_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/yourusername/routeguard/ip"
)

// ExampleNew demonstrates using the IP filter middleware with an allowlist.
func ExampleNew() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "welcome")
	})

	mw := ip.New(ip.Options{
		Allowlist: []string{"127.0.0.1"},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rec := httptest.NewRecorder()
	mw(handler).ServeHTTP(rec, req)
	fmt.Println(rec.Code)
	// Output:
	// 200
}

// ExampleNew_blocklist demonstrates using the IP filter middleware with a blocklist.
func ExampleNew_blocklist() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "welcome")
	})

	mw := ip.New(ip.Options{
		Blocklist:    []string{"192.168.0.0/16"},
		DeniedStatus: http.StatusForbidden,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.5.10:8080"
	rec := httptest.NewRecorder()
	mw(handler).ServeHTTP(rec, req)
	fmt.Println(rec.Code)
	// Output:
	// 403
}
