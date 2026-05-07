package retry

import (
	"net/http"
	"time"
)

// Options configures the retry middleware.
type Options struct {
	// MaxAttempts is the maximum number of times the handler is called.
	// Must be >= 1. Default is 3.
	MaxAttempts int

	// Delay is the wait time between attempts. Default is 100ms.
	Delay time.Duration

	// ShouldRetry determines whether a response status warrants a retry.
	// Default retries on 500, 502, 503, 504.
	ShouldRetry func(statusCode int) bool
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		MaxAttempts: 3,
		Delay:       100 * time.Millisecond,
		ShouldRetry: func(code int) bool {
			return code == http.StatusInternalServerError ||
				code == http.StatusBadGateway ||
				code == http.StatusServiceUnavailable ||
				code == http.StatusGatewayTimeout
		},
	}
}

// recorder captures the response code written by the inner handler.
type recorder struct {
	http.ResponseWriter
	code    int
	written bool
}

func (r *recorder) WriteHeader(code int) {
	if !r.written {
		r.code = code
		r.written = true
	}
	r.ResponseWriter.WriteHeader(code)
}

// New returns a middleware that retries the request up to MaxAttempts times
// when ShouldRetry returns true for the response status code.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.MaxAttempts < 1 {
		opts.MaxAttempts = DefaultOptions().MaxAttempts
	}
	if opts.ShouldRetry == nil {
		opts.ShouldRetry = DefaultOptions().ShouldRetry
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for attempt := 0; attempt < opts.MaxAttempts; attempt++ {
				rec := &recorder{ResponseWriter: w, code: http.StatusOK}
				next.ServeHTTP(rec, r)

				if !opts.ShouldRetry(rec.code) {
					return
				}

				if attempt < opts.MaxAttempts-1 && opts.Delay > 0 {
					time.Sleep(opts.Delay)
				}
			}
		})
	}
}
