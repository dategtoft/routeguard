package requestlog_test

import (
	"net/http"
	"os"

	"github.com/yourusername/routeguard/requestlog"
)

// ExampleNew demonstrates attaching the request log middleware to a handler.
func ExampleNew() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := requestlog.New(requestlog.DefaultOptions())
	http.Handle("/", mw(handler))
}

// ExampleNew_skipPaths shows how to suppress logging for specific paths such
// as health-check endpoints.
func ExampleNew_skipPaths() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	opts := requestlog.Options{
		Writer:     os.Stdout,
		SkipPaths:  []string{"/healthz", "/readyz"},
		TimeFormat: "2006-01-02T15:04:05Z",
	}

	mw := requestlog.New(opts)
	http.Handle("/api/", mw(handler))
}
