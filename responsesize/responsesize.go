// Package responsesize provides middleware that limits the size of HTTP response bodies.
package responsesize

import (
	"fmt"
	"net/http"
)

// Options configures the response size middleware.
type Options struct {
	// MaxBytes is the maximum allowed response body size in bytes.
	// Responses exceeding this limit will be truncated and a 500 status set.
	// Default: 10MB.
	MaxBytes int64

	// ErrorMessage is the message returned when the limit is exceeded.
	ErrorMessage string
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		MaxBytes:     10 * 1024 * 1024, // 10MB
		ErrorMessage: "response body too large",
	}
}

// limitedResponseWriter wraps http.ResponseWriter to cap the number of bytes written.
type limitedResponseWriter struct {
	http.ResponseWriter
	max       int64
	written   int64
	limitHit  bool
	opts      Options
}

func (lw *limitedResponseWriter) Write(p []byte) (int, error) {
	if lw.limitHit {
		return 0, fmt.Errorf("response size limit exceeded")
	}
	available := lw.max - lw.written
	if int64(len(p)) > available {
		lw.limitHit = true
		lw.ResponseWriter.Header().Set("Content-Type", "text/plain; charset=utf-8")
		lw.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		_, err := lw.ResponseWriter.Write([]byte(lw.opts.ErrorMessage))
		return 0, err
	}
	n, err := lw.ResponseWriter.Write(p)
	lw.written += int64(n)
	return n, err
}

func (lw *limitedResponseWriter) WriteHeader(code int) {
	if !lw.limitHit {
		lw.ResponseWriter.WriteHeader(code)
	}
}

// New returns middleware that limits the size of response bodies.
// If a handler attempts to write more than Options.MaxBytes, the response
// is replaced with a 500 and the configured error message.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.MaxBytes <= 0 {
		opts.MaxBytes = DefaultOptions().MaxBytes
	}
	if opts.ErrorMessage == "" {
		opts.ErrorMessage = DefaultOptions().ErrorMessage
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lw := &limitedResponseWriter{
				ResponseWriter: w,
				max:            opts.MaxBytes,
				opts:           opts,
			}
			next.ServeHTTP(lw, r)
		})
	}
}
