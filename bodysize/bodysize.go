// Package bodysize provides HTTP middleware that limits the size of incoming request bodies.
package bodysize

import (
	"net/http"
)

const (
	defaultMaxBytes int64 = 1 << 20 // 1 MB
)

// Options configures the body size middleware.
type Options struct {
	// MaxBytes is the maximum number of bytes allowed in a request body.
	// Requests exceeding this limit receive a 413 Request Entity Too Large response.
	MaxBytes int64

	// ErrorMessage is the response body sent when the limit is exceeded.
	ErrorMessage string
}

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		MaxBytes:     defaultMaxBytes,
		ErrorMessage: "Request body too large",
	}
}

// New returns middleware that enforces a maximum request body size.
// If the incoming Content-Length header already exceeds MaxBytes the request
// is rejected immediately; otherwise the body reader is wrapped with
// http.MaxBytesReader so that oversized streaming bodies are also caught.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.MaxBytes <= 0 {
		opts.MaxBytes = defaultMaxBytes
	}
	if opts.ErrorMessage == "" {
		opts.ErrorMessage = "Request body too large"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Reject early when Content-Length is known and already too large.
			if r.ContentLength > opts.MaxBytes {
				http.Error(w, opts.ErrorMessage, http.StatusRequestEntityTooLarge)
				return
			}

			// Wrap the body so streaming reads are also limited.
			r.Body = http.MaxBytesReader(w, r.Body, opts.MaxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
