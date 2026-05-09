// Package geo provides middleware for filtering HTTP requests based on geographic
// location derived from the client's IP address using a user-supplied lookup function.
package geo

import (
	"net/http"
	"strings"
)

// LookupFunc resolves a client IP address to a two-letter ISO 3166-1 alpha-2
// country code (e.g. "US", "DE"). Return an empty string when the country
// cannot be determined.
type LookupFunc func(ip string) string

// Options configures the geo middleware.
type Options struct {
	// Lookup is called with the client IP to obtain the country code.
	// Required – if nil the middleware is a no-op.
	Lookup LookupFunc

	// Allowlist, when non-empty, restricts access to the listed country codes.
	// Takes precedence over Blocklist when both are provided.
	Allowlist []string

	// Blocklist denies access for the listed country codes.
	Blocklist []string

	// DeniedCode is the HTTP status code returned when access is denied.
	// Defaults to 403.
	DeniedCode int

	// DeniedBody is the response body written when access is denied.
	// Defaults to "Forbidden".
	DeniedBody string

	// CountryHeader, when non-empty, adds the resolved country code to each
	// request under this header name before calling the next handler.
	CountryHeader string
}

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		DeniedCode: http.StatusForbidden,
		DeniedBody: "Forbidden",
	}
}

// New returns middleware that allows or denies requests based on the client's
// geographic location. When neither Allowlist nor Blocklist is set (and
// CountryHeader is also empty) the middleware is effectively a no-op.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.DeniedCode == 0 {
		opts.DeniedCode = http.StatusForbidden
	}
	if opts.DeniedBody == "" {
		opts.DeniedBody = "Forbidden"
	}

	allowSet := toSet(opts.Allowlist)
	blockSet := toSet(opts.Blocklist)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if opts.Lookup == nil {
				next.ServeHTTP(w, r)
				return
			}

			ip := clientIP(r)
			country := strings.ToUpper(opts.Lookup(ip))

			if opts.CountryHeader != "" && country != "" {
				r.Header.Set(opts.CountryHeader, country)
			}

			if len(allowSet) > 0 {
				if !allowSet[country] {
					http.Error(w, opts.DeniedBody, opts.DeniedCode)
					return
				}
			} else if len(blockSet) > 0 && blockSet[country] {
				http.Error(w, opts.DeniedBody, opts.DeniedCode)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func toSet(codes []string) map[string]bool {
	if len(codes) == 0 {
		return nil
	}
	s := make(map[string]bool, len(codes))
	for _, c := range codes {
		s[strings.ToUpper(c)] = true
	}
	return s
}

func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return strings.TrimSpace(strings.SplitN(fwd, ",", 2)[0])
	}
	host := r.RemoteAddr
	if i := strings.LastIndex(host, ":"); i != -1 {
		return host[:i]
	}
	return host
}
