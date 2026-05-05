// Package stripprefix provides HTTP middleware that strips a URL path prefix
// before passing the request to the next handler.
package stripprefix

import (
	"net/http"
	"strings"
)

// Options configures the StripPrefix middleware.
type Options struct {
	// Prefix is the path prefix to strip from incoming requests.
	Prefix string
	// RedirectOnMismatch, when true, returns a 404 for requests that do not
	// match the configured prefix instead of passing them through unchanged.
	RedirectOnMismatch bool
}

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Prefix:             "",
		RedirectOnMismatch: false,
	}
}

// New returns middleware that strips Prefix from the request URL path.
// If the request path does not start with Prefix and RedirectOnMismatch is
// true, the middleware responds with 404 Not Found.
func New(opts Options) func(http.Handler) http.Handler {
	prefix := "/" + strings.Trim(opts.Prefix, "/")
	if prefix == "/" {
		prefix = ""
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if prefix == "" {
				next.ServeHTTP(w, r)
				return
			}

			if !strings.HasPrefix(r.URL.Path, prefix) {
				if opts.RedirectOnMismatch {
					http.NotFound(w, r)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			// Strip the prefix and ensure the resulting path starts with /.
			stripped := strings.TrimPrefix(r.URL.Path, prefix)
			if stripped == "" || stripped[0] != '/' {
				stripped = "/" + stripped
			}

			// Clone the request so we do not mutate the original.
			r2 := r.Clone(r.Context())
			r2.URL.Path = stripped
			if r2.URL.RawPath != "" {
				rawStripped := strings.TrimPrefix(r2.URL.RawPath, prefix)
				if rawStripped == "" || rawStripped[0] != '/' {
					rawStripped = "/" + rawStripped
				}
				r2.URL.RawPath = rawStripped
			}

			next.ServeHTTP(w, r2)
		})
	}
}
