package signature_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/joeydtaylor/routeguard/signature"
)

// ExampleNew demonstrates protecting an endpoint with HMAC request signing.
func ExampleNew() {
	secret := "my-shared-secret"

	protected := signature.New(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "authenticated")
	}))

	// Build and sign a request.
	req := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	sig := "sha256=" + signature.Sign(secret, http.MethodGet, "/api/resource")
	req.Header.Set("X-Signature", sig)

	rec := httptest.NewRecorder()
	protected.ServeHTTP(rec, req)
	fmt.Print(rec.Body.String())
	// Output: authenticated
}

// ExampleNew_customOptions shows using a custom header and no prefix.
func ExampleNew_customOptions() {
	secret := "webhook-secret"
	opts := signature.Options{
		Header: "X-Hub-Signature-256",
		Prefix: "",
	}

	handler := signature.New(secret, opts)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "webhook accepted")
	}))

	req := httptest.NewRequest(http.MethodPost, "/webhook", nil)
	req.Header.Set("X-Hub-Signature-256", signature.Sign(secret, http.MethodPost, "/webhook"))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	fmt.Print(rec.Body.String())
	// Output: webhook accepted
}
