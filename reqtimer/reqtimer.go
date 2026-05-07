// Package reqtimer provides middleware that measures and exposes
// per-request processing duration via a response header.
package reqtimer

import (
	"fmt"
	"net/http"
	"time"
)

// Options configures the request timer middleware.
type Options struct {
	// Header is the response header name used to report elapsed time.
	// Defaults to "X-Request-Duration".
	Header string

	// Precision controls the time unit used when formatting the duration.
	// Defaults to time.Millisecond.
	Precision time.Duration
}

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Header:    "X-Request-Duration",
		Precision: time.Millisecond,
	}
}

// New returns middleware that records how long the next handler takes and
// writes the result into a response header.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.Header == "" {
		opts.Header = DefaultOptions().Header
	}
	if opts.Precision <= 0 {
		opts.Precision = DefaultOptions().Precision
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			elapsed := time.Since(start)
			units := unitLabel(opts.Precision)
			w.Header().Set(opts.Header, fmt.Sprintf("%.3f%s", float64(elapsed)/float64(opts.Precision), units))
		})
	}
}

func unitLabel(p time.Duration) string {
	switch p {
	case time.Nanosecond:
		return "ns"
	case time.Microsecond:
		return "µs"
	case time.Millisecond:
		return "ms"
	case time.Second:
		return "s"
	default:
		return ""
	}
}
