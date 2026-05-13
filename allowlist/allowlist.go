// Package allowlist provides middleware that restricts access to a set of
// allowed HTTP methods, returning 405 Method Not Allowed for anything else.
package allowlist

import (
	"net/http"
	"strings"
)

// Options holds configuration for the allowlist middleware.
type Options struct {
	// Methods is the list of HTTP methods that are permitted.
	// Defaults to [GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS].
	Methods []string

	// Message is the response body sent when a method is not allowed.
	Message string
}

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Methods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		Message: "Method Not Allowed",
	}
}

// New returns middleware that only passes through requests whose HTTP method
// is present in opts.Methods. All other methods receive a 405 response.
func New(opts Options) func(http.Handler) http.Handler {
	if len(opts.Methods) == 0 {
		opts = DefaultOptions()
	}
	if opts.Message == "" {
		opts.Message = DefaultOptions().Message
	}

	allowed := make(map[string]struct{}, len(opts.Methods))
	for _, m := range opts.Methods {
		allowed[strings.ToUpper(m)] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := allowed[r.Method]; !ok {
				http.Error(w, opts.Message, http.StatusMethodNotAllowed)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
