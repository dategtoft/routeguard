// Package methodoverride provides HTTP method override middleware.
// It allows clients to send a POST request with a header or form field
// indicating the desired HTTP method (e.g., PUT, PATCH, DELETE).
package methodoverride

import (
	"net/http"
	"strings"
)

// Options configures the method override middleware.
type Options struct {
	// Header is the HTTP header to check for the overridden method.
	// Defaults to "X-HTTP-Method-Override".
	Header string

	// FormField is the form field name to check for the overridden method.
	// Defaults to "_method".
	FormField string

	// Allowed is the set of methods that may be overridden to.
	// Defaults to PUT, PATCH, DELETE.
	Allowed []string
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Header:    "X-HTTP-Method-Override",
		FormField: "_method",
		Allowed:   []string{http.MethodPut, http.MethodPatch, http.MethodDelete},
	}
}

// New returns a middleware that overrides the HTTP method based on a header
// or form field. Only POST requests are eligible for override.
func New(opts Options) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(opts.Allowed))
	for _, m := range opts.Allowed {
		allowed[strings.ToUpper(m)] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				method := r.Header.Get(opts.Header)
				if method == "" {
					// Fall back to form field.
					method = r.FormValue(opts.FormField)
				}
				method = strings.ToUpper(method)
				if _, ok := allowed[method]; ok {
					r.Method = method
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
