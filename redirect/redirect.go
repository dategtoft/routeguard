// Package redirect provides HTTP redirect middleware for the routeguard library.
package redirect

import (
	"net/http"
	"strings"
)

// Options configures the redirect middleware.
type Options struct {
	// HTTPSOnly redirects all HTTP requests to HTTPS when true.
	HTTPSOnly bool
	// TrailingSlash controls trailing slash behaviour.
	// "add" appends a slash, "remove" strips it, "" disables.
	TrailingSlash string
	// StatusCode is the HTTP redirect status code (default 301).
	StatusCode int
}

// DefaultOptions returns sensible defaults for the redirect middleware.
func DefaultOptions() Options {
	return Options{
		HTTPSOnly:     false,
		TrailingSlash: "",
		StatusCode:    http.StatusMovedPermanently,
	}
}

// New returns a middleware that performs HTTP redirects based on the given options.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.StatusCode == 0 {
		opts.StatusCode = http.StatusMovedPermanently
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if opts.HTTPSOnly && r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https" {
				target := "https://" + r.Host + r.RequestURI
				http.Redirect(w, r, target, opts.StatusCode)
				return
			}

			path := r.URL.Path
			if opts.TrailingSlash == "add" && path != "/" && !strings.HasSuffix(path, "/") {
				http.Redirect(w, r, path+"/", opts.StatusCode)
				return
			}
			if opts.TrailingSlash == "remove" && path != "/" && strings.HasSuffix(path, "/") {
				trimmed := strings.TrimRight(path, "/")
				http.Redirect(w, r, trimmed, opts.StatusCode)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
