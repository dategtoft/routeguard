// Package nocache provides middleware that sets response headers to prevent
// clients and proxies from caching responses.
package nocache

import "net/http"

// Options configures the nocache middleware.
type Options struct {
	// Pragma sets the legacy Pragma: no-cache header (default: true).
	Pragma bool
	// Expires sets Expires: 0 header (default: true).
	Expires bool
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Pragma:  true,
		Expires: true,
	}
}

// New returns middleware that disables caching by setting appropriate
// Cache-Control, Pragma, and Expires response headers.
func New(opts Options) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
			if opts.Pragma {
				h.Set("Pragma", "no-cache")
			}
			if opts.Expires {
				h.Set("Expires", "0")
			}
			next.ServeHTTP(w, r)
		})
	}
}
