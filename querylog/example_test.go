package querylog_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/patrickward/routeguard/querylog"
)

func ExampleNew() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	opts := querylog.DefaultOptions()
	opts.Logger = log.New(os.Stdout, "", 0)

	mw := querylog.New(opts)(handler)

	req := httptest.NewRequest(http.MethodGet, "/search?q=gopher", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)
}

func ExampleNew_redactKeys() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	opts := querylog.DefaultOptions()
	opts.Logger = log.New(os.Stdout, "", 0)
	opts.RedactKeys = []string{"api_key", "token"}

	mw := querylog.New(opts)(handler)

	req := httptest.NewRequest(http.MethodGet, "/data?api_key=secret&limit=10", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)
}
