// Package realip provides middleware that resolves the real client IP
// from trusted reverse-proxy headers such as X-Forwarded-For and X-Real-IP.
package realip

import (
	"net"
	"net/http"
	"strings"
)

// Options configures the realip middleware.
type Options struct {
	// TrustedProxies is the list of CIDR ranges considered trusted.
	// Only headers from requests arriving via these ranges are honoured.
	// Defaults to loopback and private ranges.
	TrustedProxies []string

	// Headers is the ordered list of headers to inspect.
	// The first non-empty value wins.
	// Defaults to ["X-Forwarded-For", "X-Real-IP"].
	Headers []string
}

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		TrustedProxies: []string{
			"127.0.0.0/8",
			"10.0.0.0/8",
			"172.16.0.0/12",
			"192.168.0.0/16",
			"::1/128",
		},
		Headers: []string{"X-Forwarded-For", "X-Real-IP"},
	}
}

// New returns middleware that rewrites r.RemoteAddr with the real client IP
// when the direct connection comes from a trusted proxy.
func New(opts Options) func(http.Handler) http.Handler {
	nets := parseCIDRs(opts.TrustedProxies)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ip, ok := resolve(r, nets, opts.Headers); ok {
				// Preserve the original port if present.
				_, port, err := net.SplitHostPort(r.RemoteAddr)
				if err == nil {
					r.RemoteAddr = net.JoinHostPort(ip, port)
				} else {
					r.RemoteAddr = ip
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func resolve(r *http.Request, trusted []*net.IPNet, headers []string) (string, bool) {
	directIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		directIP = r.RemoteAddr
	}
	if !isTrusted(net.ParseIP(directIP), trusted) {
		return "", false
	}
	for _, h := range headers {
		val := r.Header.Get(h)
		if val == "" {
			continue
		}
		// X-Forwarded-For may be a comma-separated list; take the first entry.
		ip := strings.TrimSpace(strings.SplitN(val, ",", 2)[0])
		if net.ParseIP(ip) != nil {
			return ip, true
		}
	}
	return "", false
}

func isTrusted(ip net.IP, nets []*net.IPNet) bool {
	if ip == nil {
		return false
	}
	for _, n := range nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

func parseCIDRs(cidrs []string) []*net.IPNet {
	var out []*net.IPNet
	for _, c := range cidrs {
		_, n, err := net.ParseCIDR(c)
		if err == nil {
			out = append(out, n)
		}
	}
	return out
}
