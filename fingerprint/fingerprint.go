// Package fingerprint provides middleware that generates a request fingerprint
// based on configurable attributes (IP, User-Agent, headers) and stores it in
// the request context for downstream handlers to consume.
package fingerprint

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
)

type contextKey struct{}

// Options configures the fingerprint middleware.
type Options struct {
	// Header is the response header name to write the fingerprint into.
	// Defaults to "X-Request-Fingerprint".
	Header string

	// IncludeIP includes the remote IP in the fingerprint when true.
	IncludeIP bool

	// IncludeUserAgent includes the User-Agent header in the fingerprint when true.
	IncludeUserAgent bool

	// ExtraHeaders lists additional request headers to fold into the fingerprint.
	ExtraHeaders []string
}

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Header:           "X-Request-Fingerprint",
		IncludeIP:        true,
		IncludeUserAgent: true,
	}
}

// FromContext retrieves the fingerprint stored by the middleware.
// Returns an empty string if none is present.
func FromContext(ctx context.Context) string {
	v, _ := ctx.Value(contextKey{}).(string)
	return v
}

// New returns middleware that computes a SHA-256 fingerprint for every request
// and exposes it via the response header and request context.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.Header == "" {
		opts.Header = "X-Request-Fingerprint"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fp := compute(r, opts)
			w.Header().Set(opts.Header, fp)
			r = r.WithContext(context.WithValue(r.Context(), contextKey{}, fp))
			next.ServeHTTP(w, r)
		})
	}
}

func compute(r *http.Request, opts Options) string {
	var parts []string

	if opts.IncludeIP {
		ip := r.RemoteAddr
		if idx := strings.LastIndex(ip, ":"); idx != -1 {
			ip = ip[:idx]
		}
		parts = append(parts, ip)
	}

	if opts.IncludeUserAgent {
		parts = append(parts, r.UserAgent())
	}

	for _, h := range opts.ExtraHeaders {
		parts = append(parts, r.Header.Get(h))
	}

	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return fmt.Sprintf("%x", sum)
}
