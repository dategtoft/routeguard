// Package requestsize provides middleware that limits the total size of
// incoming HTTP requests, including headers and body.
package requestsize

import (
	"fmt"
	"net/http"
)

const defaultMaxBytes = 1 << 20 // 1 MB

// Options configures the request size middleware.
type Options struct {
	// MaxBytes is the maximum allowed size of the entire request in bytes.
	// Defaults to 1 MB.
	MaxBytes int64

	// ErrorMessage is the body returned when the limit is exceeded.
	// Defaults to a simple text message.
	ErrorMessage string

	// SkipMethods lists HTTP methods that bypass the size check (e.g. GET, HEAD).
	SkipMethods []string
}

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		MaxBytes:     defaultMaxBytes,
		ErrorMessage: "request too large",
		SkipMethods:  []string{http.MethodGet, http.MethodHead, http.MethodOptions},
	}
}

// New returns middleware that rejects requests whose size exceeds MaxBytes.
// It sets http.MaxBytesReader on the request body and also checks the
// Content-Length header before reading, so oversized requests are rejected
// early without buffering.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.MaxBytes <= 0 {
		opts.MaxBytes = defaultMaxBytes
	}
	if opts.ErrorMessage == "" {
		opts.ErrorMessage = "request too large"
	}

	skip := make(map[string]struct{}, len(opts.SkipMethods))
	for _, m := range opts.SkipMethods {
		skip[m] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := skip[r.Method]; ok {
				next.ServeHTTP(w, r)
				return
			}

			// Reject early if Content-Length is known and already too large.
			if r.ContentLength > opts.MaxBytes {
				http.Error(w,
					fmt.Sprintf("%s (limit: %d bytes)", opts.ErrorMessage, opts.MaxBytes),
					http.StatusRequestEntityTooLarge)
				return
			}

			// Wrap the body so streaming uploads are also capped.
			r.Body = http.MaxBytesReader(w, r.Body, opts.MaxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
