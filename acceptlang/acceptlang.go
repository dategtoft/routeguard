// Package acceptlang provides middleware that parses the Accept-Language header
// and stores the preferred language in the request context.
package acceptlang

import (
	"context"
	"net/http"
	"strings"
)

type contextKey struct{}

// Options configures the Accept-Language middleware.
type Options struct {
	// Supported is the list of language tags the application supports.
	// If empty, all languages are accepted.
	Supported []string
	// Default is the language to use when no match is found.
	Default string
	// Header is the response header to set with the resolved language.
	// Leave empty to skip setting a response header.
	Header string
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Default: "en",
		Header:  "Content-Language",
	}
}

// FromContext retrieves the resolved language from ctx.
// Returns an empty string if not present.
func FromContext(ctx context.Context) string {
	v, _ := ctx.Value(contextKey{}).(string)
	return v
}

// New returns middleware that parses Accept-Language and stores the best
// matching language in the request context.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.Default == "" {
		opts.Default = "en"
	}
	supported := make(map[string]bool, len(opts.Supported))
	for _, s := range opts.Supported {
		supported[strings.ToLower(s)] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lang := resolve(r.Header.Get("Accept-Language"), supported, opts.Default)
			ctx := context.WithValue(r.Context(), contextKey{}, lang)
			if opts.Header != "" {
				w.Header().Set(opts.Header, lang)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// resolve picks the best language from the Accept-Language header value.
func resolve(header string, supported map[string]bool, fallback string) string {
	if header == "" {
		return fallback
	}
	for _, part := range strings.Split(header, ",") {
		tag := strings.ToLower(strings.TrimSpace(strings.SplitN(part, ";", 2)[0]))
		if tag == "" {
			continue
		}
		if len(supported) == 0 || supported[tag] {
			return tag
		}
		// Try base language (e.g. "en" from "en-US")
		if idx := strings.Index(tag, "-"); idx > 0 {
			base := tag[:idx]
			if len(supported) == 0 || supported[base] {
				return base
			}
		}
	}
	return fallback
}
