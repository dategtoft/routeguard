package geo_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/yourusername/routeguard/geo"
)

// ExampleNew demonstrates an allowlist that only permits US and CA traffic.
func ExampleNew() {
	lookup := func(ip string) string {
		// Replace with a real GeoIP database call.
		return "US"
	}

	opts := geo.DefaultOptions()
	opts.Lookup = lookup
	opts.Allowlist = []string{"US", "CA"}
	opts.CountryHeader = "X-Country-Code"

	mw := geo.New(opts)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from %s", r.Header.Get("X-Country-Code"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	fmt.Println(rec.Code)
	// Output:
	// 200
}

// ExampleNew_blocklist demonstrates blocking specific countries.
func ExampleNew_blocklist() {
	lookup := func(ip string) string {
		return "KP"
	}

	opts := geo.DefaultOptions()
	opts.Lookup = lookup
	opts.Blocklist = []string{"KP", "CU"}

	mw := geo.New(opts)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	fmt.Println(rec.Code)
	// Output:
	// 403
}
