// Package contenttype provides middleware that enforces or sets Content-Type
// headers on HTTP requests and responses.
package contenttype

import (
	"net/http"
	"strings"
)

// Options configures the Content-Type middleware.
type Options struct {
	// Allowed is the list of accepted request Content-Type values.
	// If empty, all content types are permitted.
	Allowed []string

	// ResponseType is the Content-Type value to set on every response.
	// If empty, the response Content-Type is left unchanged.
	ResponseType string

	// SkipMethods lists HTTP methods whose request body content type is not
	// checked (e.g. GET, HEAD). Defaults to GET, HEAD, OPTIONS.
	SkipMethods []string
}

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		SkipMethods: []string{http.MethodGet, http.MethodHead, http.MethodOptions},
	}
}

// New returns middleware that validates request Content-Type headers and
// optionally sets a fixed Content-Type on every response.
func New(opts Options) func(http.Handler) http.Handler {
	skip := make(map[string]struct{}, len(opts.SkipMethods))
	for _, m := range opts.SkipMethods {
		skip[strings.ToUpper(m)] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := skip[r.Method]; !ok && len(opts.Allowed) > 0 {
				ct := r.Header.Get("Content-Type")
				// Strip parameters (e.g. charset) before comparison.
				if idx := strings.Index(ct, ";"); idx != -1 {
					ct = strings.TrimSpace(ct[:idx])
				}
				if !isAllowed(ct, opts.Allowed) {
					http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
					return
				}
			}

			if opts.ResponseType != "" {
				w.Header().Set("Content-Type", opts.ResponseType)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isAllowed(ct string, allowed []string) bool {
	for _, a := range allowed {
		if strings.EqualFold(ct, a) {
			return true
		}
	}
	return false
}
