package responsesize_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/patrickward/routeguard/responsesize"
)

// ExampleNew demonstrates basic usage of the responsesize middleware
// with the default 10 MB cap.
func ExampleNew() {
	handler := responsesize.New(responsesize.DefaultOptions())(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "Hello, world!")
		}),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Body.String())
	// Output:
	// 200
	// Hello, world!
}

// ExampleNew_limited shows how to enforce a tight response size limit.
func ExampleNew_limited() {
	bigBody := strings.Repeat("x", 512)

	handler := responsesize.New(responsesize.Options{
		MaxBytes:     256,
		ErrorMessage: "response body too large",
	})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, bigBody)
		}),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Body.String())
	// Output:
	// 500
	// response body too large
}
