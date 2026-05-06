package sanitize_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/joeychilson/routeguard/sanitize"
)

// ExampleNew demonstrates the sanitize middleware with default options,
// which escapes HTML entities and removes null bytes from query parameters.
func ExampleNew() {
	handler := sanitize.New(sanitize.DefaultOptions())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, r.URL.Query().Get("q"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/?q=%3Cscript%3E", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	fmt.Print(rec.Body.String())
	// Output:
	// &lt;script&gt;
}

// ExampleNew_trimSpace shows how to enable whitespace trimming.
func ExampleNew_trimSpace() {
	opts := sanitize.DefaultOptions()
	opts.TrimSpace = true

	handler := sanitize.New(opts)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%q", r.URL.Query().Get("name"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/?name=+Alice+", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	fmt.Print(rec.Body.String())
	// Output:
	// "Alice"
}
