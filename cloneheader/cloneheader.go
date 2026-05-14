// Package cloneheader provides middleware that copies values from one
// request header into another before passing the request downstream.
// This is useful when a reverse proxy rewrites headers and you need to
// preserve or alias them under a canonical name.
package cloneheader

import "net/http"

// Options configures the CloneHeader middleware.
type Options struct {
	// Rules is an ordered list of (Source, Destination) pairs.
	// For each rule the value of Source header is copied into Destination.
	// If Source is absent the rule is silently skipped.
	Rules []Rule
	// Overwrite controls whether an already-present Destination header is
	// replaced. Defaults to false (skip if destination already set).
	Overwrite bool
}

// Rule describes a single header-clone operation.
type Rule struct {
	Source      string
	Destination string
}

// DefaultOptions returns a zero-value Options with no rules.
func DefaultOptions() Options {
	return Options{}
}

// New returns middleware that clones request headers according to opts.
func New(opts Options, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Work on a shallow clone so we don't mutate the original header map
		// in place when Overwrite is false and the destination already exists.
		newHeader := r.Header.Clone()

		for _, rule := range opts.Rules {
			src := r.Header.Get(rule.Source)
			if src == "" {
				continue
			}
			if !opts.Overwrite && newHeader.Get(rule.Destination) != "" {
				continue
			}
			newHeader.Set(rule.Destination, src)
		}

		r2 := r.Clone(r.Context())
		r2.Header = newHeader
		next.ServeHTTP(w, r2)
	})
}
