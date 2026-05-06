package contenttype_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/username/routeguard/contenttype"
)

// ExampleNew demonstrates enforcing JSON request bodies.
func ExampleNew() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := contenttype.New(contenttype.Options{
		Allowed:      []string{"application/json"},
		ResponseType: "application/json",
		SkipMethods:  []string{http.MethodGet, http.MethodHead},
	})

	ts := httptest.NewServer(mw(handler))
	defer ts.Close()

	resp, _ := http.Post(ts.URL, "application/json", strings.NewReader(`{}`))
	fmt.Println(resp.StatusCode)
	// Output: 200
}

// ExampleNew_rejected shows a request blocked due to wrong Content-Type.
func ExampleNew_rejected() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := contenttype.New(contenttype.Options{
		Allowed:     []string{"application/json"},
		SkipMethods: []string{http.MethodGet},
	})

	ts := httptest.NewServer(mw(handler))
	defer ts.Close()

	resp, _ := http.Post(ts.URL, "text/xml", strings.NewReader(`<a/>`))
	fmt.Println(resp.StatusCode)
	// Output: 415
}
