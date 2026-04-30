package csrf_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/yourusername/routeguard/csrf"
)

// ExampleNew demonstrates using the CSRF middleware with default options.
func ExampleNew() {
	protected := csrf.New(csrf.DefaultOptions())(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "hello")
		}),
	)

	// GET request — token cookie is issued automatically.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	protected.ServeHTTP(rec, req)
	fmt.Println(rec.Code)

	// Output:
	// 200
}

// ExampleNew_customOptions shows how to configure a custom header and cookie name.
func ExampleNew_customOptions() {
	opts := csrf.Options{
		TokenHeader: "X-My-CSRF-Token",
		CookieName:  "my_csrf",
		TokenLength: 16,
		SafeMethods: []string{http.MethodGet, http.MethodHead},
	}

	protected := csrf.New(opts)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "ok")
		}),
	)

	const token = "my-secret-token"
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/submit", nil)
	req.AddCookie(&http.Cookie{Name: "my_csrf", Value: token})
	req.Header.Set("X-My-CSRF-Token", token)
	protected.ServeHTTP(rec, req)
	fmt.Println(rec.Code)

	// Output:
	// 200
}
