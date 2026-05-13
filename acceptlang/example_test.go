package acceptlang_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/joeydtaylor/routeguard/acceptlang"
)

func ExampleNew() {
	opts := acceptlang.DefaultOptions()
	opts.Supported = []string{"en", "fr", "de"}

	mw := acceptlang.New(opts)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := acceptlang.FromContext(r.Context())
		fmt.Fprintf(w, "language: %s", lang)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	fmt.Println(rec.Body.String())
	// Output: language: fr
}

func ExampleNew_customOptions() {
	opts := acceptlang.Options{
		Supported: []string{"en", "es"},
		Default:   "en",
		Header:    "", // do not set a response header
	}

	mw := acceptlang.New(opts)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := acceptlang.FromContext(r.Context())
		fmt.Fprintf(w, "language: %s", lang)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "zh, es;q=0.9")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	fmt.Println(rec.Body.String())
	// Output: language: es
}
