// Package sanitize provides HTTP middleware that strips or escapes
// potentially dangerous characters from request query parameters and
// form fields before the request reaches the next handler.
package sanitize

import (
	"html"
	"net/http"
	"net/url"
	"strings"
)

// Options configures the sanitize middleware.
type Options struct {
	// EscapeHTML replaces <, >, &, ' and " with their HTML entities.
	// Defaults to true.
	EscapeHTML bool

	// StripNullBytes removes null bytes (\x00) from values.
	// Defaults to true.
	StripNullBytes bool

	// TrimSpace trims leading and trailing whitespace from values.
	// Defaults to false.
	TrimSpace bool
}

// DefaultOptions returns an Options with safe defaults.
func DefaultOptions() Options {
	return Options{
		EscapeHTML:     true,
		StripNullBytes: true,
		TrimSpace:      false,
	}
}

// New returns middleware that sanitizes query parameters and form fields
// according to the provided Options.
func New(opts Options) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Sanitize query parameters.
			if raw := r.URL.RawQuery; raw != "" {
				parsed, err := url.ParseQuery(raw)
				if err == nil {
					r.URL.RawQuery = sanitizeValues(parsed, opts).Encode()
				}
			}

			// Sanitize already-parsed form values if present.
			if r.Form != nil {
				r.Form = sanitizeValues(r.Form, opts)
			}
			if r.PostForm != nil {
				r.PostForm = sanitizeValues(r.PostForm, opts)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// NewDefault returns middleware using DefaultOptions. It is a convenience
// wrapper around New(DefaultOptions()) for the common case.
func NewDefault() func(http.Handler) http.Handler {
	return New(DefaultOptions())
}

func sanitizeValues(vals url.Values, opts Options) url.Values {
	out := make(url.Values, len(vals))
	for key, slice := range vals {
		cleaned := make([]string, len(slice))
		for i, v := range slice {
			cleaned[i] = sanitizeString(v, opts)
		}
		out[key] = cleaned
	}
	return out
}

func sanitizeString(s string, opts Options) string {
	if opts.StripNullBytes {
		s = strings.ReplaceAll(s, "\x00", "")
	}
	if opts.TrimSpace {
		s = strings.TrimSpace(s)
	}
	if opts.EscapeHTML {
		s = html.EscapeString(s)
	}
	return s
}
