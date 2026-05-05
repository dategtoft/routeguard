// Package ip provides middleware for IP-based access control (allowlist/blocklist).
package ip

import (
	"net"
	"net/http"
	"strings"
)

// Options configures the IP filter middleware.
type Options struct {
	// Allowlist is a list of IPs or CIDR ranges that are permitted.
	// If non-empty, only these IPs are allowed.
	Allowlist []string
	// Blocklist is a list of IPs or CIDR ranges that are denied.
	Blocklist []string
	// DeniedStatus is the HTTP status code returned when access is denied.
	// Defaults to 403.
	DeniedStatus int
	// TrustProxy indicates whether to trust X-Forwarded-For headers.
	TrustProxy bool
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		DeniedStatus: http.StatusForbidden,
		TrustProxy:   false,
	}
}

type ipFilter struct {
	opts      Options
	allowNets []*net.IPNet
	blockNets []*net.IPNet
}

// New returns an IP filter middleware based on the provided Options.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.DeniedStatus == 0 {
		opts.DeniedStatus = http.StatusForbidden
	}
	f := &ipFilter{opts: opts}
	f.allowNets = parseCIDRs(opts.Allowlist)
	f.blockNets = parseCIDRs(opts.Blocklist)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r, opts.TrustProxy)
			parsed := net.ParseIP(ip)
			if parsed == nil || f.isDenied(parsed) {
				http.Error(w, http.StatusText(opts.DeniedStatus), opts.DeniedStatus)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (f *ipFilter) isDenied(ip net.IP) bool {
	if len(f.allowNets) > 0 {
		for _, n := range f.allowNets {
			if n.Contains(ip) {
				return false
			}
		}
		return true
	}
	for _, n := range f.blockNets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// parseCIDRs parses a list of IP addresses or CIDR ranges into []*net.IPNet.
// Plain IP addresses (without a prefix length) are treated as /32 for IPv4
// or /128 for IPv6.
func parseCIDRs(entries []string) []*net.IPNet {
	var nets []*net.IPNet
	for _, e := range entries {
		if !strings.Contains(e, "/") {
			// Use /128 for IPv6 addresses, /32 for IPv4.
			if strings.Contains(e, ":") {
				e += "/128"
			} else {
				e += "/32"
			}
		}
		_, n, err := net.ParseCIDR(e)
		if err == nil {
			nets = append(nets, n)
		}
	}
	return nets
}

func clientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			return strings.TrimSpace(strings.SplitN(fwd, ",", 2)[0])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
