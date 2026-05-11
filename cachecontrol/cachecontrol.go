// Package cachecontrol provides HTTP middleware for setting Cache-Control headers.
package cachecontrol

import (
	"fmt"
	"net/http"
	"strings"
)

// Options configures the Cache-Control middleware.
type Options struct {
	// MaxAge sets the max-age directive in seconds. Ignored when Private is true.
	MaxAge int
	// Private marks the response as private (not cacheable by shared caches).
	Private bool
	// NoStore instructs caches not to store the response.
	NoStore bool
	// NoCache forces revalidation on each request.
	NoCache bool
	// Immutable adds the immutable directive (useful for versioned assets).
	Immutable bool
	// MustRevalidate adds the must-revalidate directive.
	MustRevalidate bool
	// SkipMethods is a list of HTTP methods to skip (e.g. ["POST", "PUT"]).
	SkipMethods []string
}

// DefaultOptions returns a sensible default: public, 60-second max-age.
func DefaultOptions() Options {
	return Options{
		MaxAge: 60,
	}
}

// New returns middleware that sets the Cache-Control header on responses.
func New(opts Options) func(http.Handler) http.Handler {
	skip := make(map[string]struct{}, len(opts.SkipMethods))
	for _, m := range opts.SkipMethods {
		skip[strings.ToUpper(m)] = struct{}{}
	}

	header := build(opts)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, shouldSkip := skip[r.Method]; !shouldSkip {
				w.Header().Set("Cache-Control", header)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// build constructs the Cache-Control header value from Options.
func build(opts Options) string {
	var parts []string

	if opts.NoStore {
		return "no-store"
	}

	if opts.NoCache {
		parts = append(parts, "no-cache")
	}

	if opts.Private {
		parts = append(parts, "private")
	} else {
		parts = append(parts, "public")
		if opts.MaxAge > 0 {
			parts = append(parts, fmt.Sprintf("max-age=%d", opts.MaxAge))
		}
	}

	if opts.Immutable {
		parts = append(parts, "immutable")
	}

	if opts.MustRevalidate {
		parts = append(parts, "must-revalidate")
	}

	return strings.Join(parts, ", ")
}
