// Package slowdown provides a middleware that introduces an artificial delay
// on responses, useful for simulating latency in development and testing.
package slowdown

import (
	"net/http"
	"time"
)

// Options holds configuration for the slowdown middleware.
type Options struct {
	// Delay is the duration to wait before passing the request to the next handler.
	Delay time.Duration
	// OnlyPaths restricts the delay to specific URL paths. If empty, all paths are delayed.
	OnlyPaths []string
}

// DefaultOptions returns Options with a 500ms delay applied to all paths.
func DefaultOptions() Options {
	return Options{
		Delay: 500 * time.Millisecond,
	}
}

// New returns a middleware that delays each request by the configured duration.
func New(opts Options) func(http.Handler) http.Handler {
	pathSet := make(map[string]struct{}, len(opts.OnlyPaths))
	for _, p := range opts.OnlyPaths {
		pathSet[p] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if shouldDelay(r.URL.Path, pathSet) {
				select {
				case <-time.After(opts.Delay):
				case <-r.Context().Done():
					http.Error(w, "request cancelled", http.StatusServiceUnavailable)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func shouldDelay(path string, pathSet map[string]struct{}) bool {
	if len(pathSet) == 0 {
		return true
	}
	_, ok := pathSet[path]
	return ok
}
