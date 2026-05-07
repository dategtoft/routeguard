// Package expiry provides middleware that sets Cache-Control and Expires
// headers on responses to control browser and proxy caching lifetimes.
package expiry

import (
	"fmt"
	"net/http"
	"time"
)

// Options configures the expiry middleware.
type Options struct {
	// MaxAge is how long the response may be cached.
	MaxAge time.Duration

	// Immutable adds the immutable directive (useful for versioned assets).
	Immutable bool

	// NoStore disables caching entirely (overrides MaxAge).
	NoStore bool

	// Private marks the response as private (not cacheable by shared caches).
	Private bool
}

// DefaultOptions returns Options with a 5-minute max age.
func DefaultOptions() Options {
	return Options{
		MaxAge: 5 * time.Minute,
	}
}

// New returns middleware that applies Cache-Control and Expires headers.
func New(opts Options) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if opts.NoStore {
				w.Header().Set("Cache-Control", "no-store")
				w.Header().Set("Pragma", "no-cache")
				next.ServeHTTP(w, r)
				return
			}

			scope := "public"
			if opts.Private {
				scope = "private"
			}

			seconds := int(opts.MaxAge.Seconds())
			directive := fmt.Sprintf("%s, max-age=%d", scope, seconds)
			if opts.Immutable {
				directive += ", immutable"
			}

			w.Header().Set("Cache-Control", directive)
			w.Header().Set("Expires", time.Now().UTC().Add(opts.MaxAge).Format(http.TimeFormat))

			next.ServeHTTP(w, r)
		})
	}
}
