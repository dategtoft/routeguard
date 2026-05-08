package fingerprint_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/joeydtaylor/routeguard/fingerprint"
)

// ExampleNew demonstrates basic fingerprint middleware usage.
func ExampleNew() {
	handler := fingerprint.New(fingerprint.DefaultOptions())(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fp := fingerprint.FromContext(r.Context())
			fmt.Fprintln(w, "fingerprint:", fp)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.0.1:4321"
	req.Header.Set("User-Agent", "example-agent")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	fp := rec.Header().Get("X-Request-Fingerprint")
	fmt.Println("header length:", len(fp))
	// Output:
	// header length: 64
}

// ExampleNew_extraHeaders demonstrates including a custom header in the fingerprint.
func ExampleNew_extraHeaders() {
	opts := fingerprint.DefaultOptions()
	opts.ExtraHeaders = []string{"X-Tenant-ID"}

	handler := fingerprint.New(opts)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:8080"
	req.Header.Set("X-Tenant-ID", "tenant-abc")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	fmt.Println("status:", rec.Code)
	// Output:
	// status: 200
}
