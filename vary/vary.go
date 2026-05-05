// Package vary provides middleware that appends headers to the Vary response header.
package vary

import (
	"net/http"
	"strings"
)

// Options configures the Vary middleware.
type Options struct {
	// Headers is the list of request headers to add to the Vary response header.
	Headers []string
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Headers: []string{"Accept-Encoding"},
	}
}

// New returns middleware that appends the configured headers to the Vary
// response header. Multiple calls to New can be chained; headers are merged
// rather than overwritten.
func New(opts Options) func(http.Handler) http.Handler {
	// Deduplicate the configured headers (case-insensitive).
	seen := make(map[string]struct{}, len(opts.Headers))
	headers := make([]string, 0, len(opts.Headers))
	for _, h := range opts.Headers {
		key := strings.ToLower(h)
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			headers = append(headers, http.CanonicalHeaderKey(h))
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			appendVary(w, headers)
			next.ServeHTTP(w, r)
		})
	}
}

// appendVary merges headers into the existing Vary value, skipping duplicates
// and honouring the wildcard (*) short-circuit.
func appendVary(w http.ResponseWriter, headers []string) {
	existing := w.Header().Get("Vary")
	if existing == "*" {
		return
	}

	// Build a set of already-present vary tokens.
	present := make(map[string]struct{})
	if existing != "" {
		for _, part := range strings.Split(existing, ",") {
			present[strings.ToLower(strings.TrimSpace(part))] = struct{}{}
		}
	}

	for _, h := range headers {
		if h == "*" {
			w.Header().Set("Vary", "*")
			return
		}
		if _, ok := present[strings.ToLower(h)]; !ok {
			present[strings.ToLower(h)] = struct{}{}
			if existing == "" {
				existing = h
			} else {
				existing += ", " + h
			}
		}
	}

	w.Header().Set("Vary", existing)
}
