// Package proxy provides HTTP reverse proxy middleware with optional path stripping.
package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// Options configures the reverse proxy middleware.
type Options struct {
	// Target is the base URL to proxy requests to (required).
	Target string

	// StripPrefix removes this prefix from the request path before forwarding.
	StripPrefix string

	// Timeout is the maximum duration for a proxied request.
	Timeout time.Duration

	// ModifyRequest allows callers to mutate the outbound request.
	ModifyRequest func(*http.Request)
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions(target string) Options {
	return Options{
		Target:  target,
		Timeout: 30 * time.Second,
	}
}

// New returns middleware that reverse-proxies requests to the configured target.
// It panics if Target is empty or cannot be parsed as a URL.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.Target == "" {
		panic("proxy: Target must not be empty")
	}

	targetURL, err := url.Parse(opts.Target)
	if err != nil {
		panic("proxy: invalid Target URL: " + err.Error())
	}

	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second
	}

	rp := httputil.NewSingleHostReverseProxy(targetURL)

	originalDirector := rp.Director
	rp.Director = func(req *http.Request) {
		originalDirector(req)
		if opts.StripPrefix != "" {
			req.URL.Path = strings.TrimPrefix(req.URL.Path, opts.StripPrefix)
			if req.URL.Path == "" {
				req.URL.Path = "/"
			}
			req.URL.RawPath = strings.TrimPrefix(req.URL.RawPath, opts.StripPrefix)
		}
		req.Header.Set("X-Forwarded-Host", req.Host)
		if opts.ModifyRequest != nil {
			opts.ModifyRequest(req)
		}
	}

	rp.Transport = &http.Transport{
		ResponseHeaderTimeout: opts.Timeout,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rp.ServeHTTP(w, r)
		})
	}
}
