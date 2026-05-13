// Package dnsblocklist provides middleware that blocks requests from hosts
// whose reverse-DNS lookup matches a configurable list of blocked domains.
package dnsblocklist

import (
	"net"
	"net/http"
	"strings"
)

// Options configures the DNS blocklist middleware.
type Options struct {
	// BlockedDomains is a list of domain suffixes to block (e.g. "example.com").
	BlockedDomains []string

	// StatusCode is returned when a request is blocked. Defaults to 403.
	StatusCode int

	// Message is the response body when blocked.
	Message string

	// LookupAddr overrides the DNS lookup function (useful for testing).
	LookupAddr func(addr string) ([]string, error)
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		StatusCode: http.StatusForbidden,
		Message:    "Forbidden",
		LookupAddr: net.LookupAddr,
	}
}

// New returns middleware that blocks requests whose remote IP reverse-resolves
// to a hostname matching any of the configured blocked domain suffixes.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.StatusCode == 0 {
		opts.StatusCode = http.StatusForbidden
	}
	if opts.Message == "" {
		opts.Message = "Forbidden"
	}
	if opts.LookupAddr == nil {
		opts.LookupAddr = net.LookupAddr
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			if ip != "" && len(opts.BlockedDomains) > 0 {
				hosts, err := opts.LookupAddr(ip)
				if err == nil {
					for _, host := range hosts {
						host = strings.ToLower(strings.TrimSuffix(host, "."))
						for _, blocked := range opts.BlockedDomains {
							blocked = strings.ToLower(blocked)
							if host == blocked || strings.HasSuffix(host, "."+blocked) {
								http.Error(w, opts.Message, opts.StatusCode)
								return
							}
						}
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
