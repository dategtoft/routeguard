// Package reqbody provides middleware that buffers and optionally logs
// the raw request body, making it available for downstream handlers.
package reqbody

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

type contextKey struct{}

// Options configures the request body middleware.
type Options struct {
	// MaxBytes is the maximum number of bytes to buffer. 0 means no limit.
	MaxBytes int64
	// OnRead is an optional callback invoked with the raw body bytes.
	OnRead func(r *http.Request, body []byte)
}

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		MaxBytes: 1 << 20, // 1 MiB
	}
}

// FromContext retrieves the buffered request body from the context.
// Returns nil if no body was stored.
func FromContext(ctx context.Context) []byte {
	v, _ := ctx.Value(contextKey{}).([]byte)
	return v
}

// New returns middleware that reads and buffers the request body so that
// downstream handlers can access it via FromContext.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.MaxBytes == 0 {
		opts.MaxBytes = DefaultOptions().MaxBytes
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body == nil {
				next.ServeHTTP(w, r)
				return
			}
			limited := io.LimitReader(r.Body, opts.MaxBytes)
			data, err := io.ReadAll(limited)
			r.Body.Close()
			if err != nil {
				http.Error(w, "failed to read request body", http.StatusInternalServerError)
				return
			}
			// Restore body so downstream handlers can read it normally.
			r.Body = io.NopCloser(bytes.NewReader(data))
			if opts.OnRead != nil {
				opts.OnRead(r, data)
			}
			ctx := context.WithValue(r.Context(), contextKey{}, data)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
