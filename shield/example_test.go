package shield_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/joeydtaylor/routeguard/shield"
)

func ExampleNew() {
	handler := shield.New(shield.DefaultOptions())(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	fmt.Println(rec.Header().Get("X-Frame-Options"))
	// Output: DENY
}

func ExampleNew_customOptions() {
	opts := shield.DefaultOptions()
	opts.XFrameOptions = "SAMEORIGIN"
	opts.PermissionsPolicy = "geolocation=()"

	handler := shield.New(opts)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	fmt.Println(rec.Header().Get("X-Frame-Options"))
	fmt.Println(rec.Header().Get("Permissions-Policy"))
	// Output:
	// SAMEORIGIN
	// geolocation=()
}
