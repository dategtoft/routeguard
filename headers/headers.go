// Package headers provides middleware for setting common security and custom HTTP response headers.
package headers

import "net/http"

// Options configures the headers middleware.
type Options struct {
	// SecurityHeaders enables common security headers (CSP, HSTS, X-Frame-Options, etc.).
	SecurityHeaders bool
	// Custom is a map of additional headers to set on every response.
	Custom map[string]string
	// HSTSMaxAge sets the max-age for the Strict-Transport-Security header (seconds).
	// Only applied when SecurityHeaders is true. Defaults to 31536000 (1 year).
	HSTSMaxAge int
}

// DefaultOptions returns an Options with security headers enabled and a 1-year HSTS max-age.
func DefaultOptions() Options {
	return Options{
		SecurityHeaders: true,
		HSTSMaxAge:      31536000,
		Custom:          map[string]string{},
	}
}

// New returns middleware that sets HTTP response headers according to opts.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.HSTSMaxAge == 0 {
		opts.HSTSMaxAge = 31536000
	}
	if opts.Custom == nil {
		opts.Custom = map[string]string{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if opts.SecurityHeaders {
				setSecurityHeaders(w, opts.HSTSMaxAge)
			}
			for k, v := range opts.Custom {
				w.Header().Set(k, v)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func setSecurityHeaders(w http.ResponseWriter, hstsMaxAge int) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	w.Header().Set("Content-Security-Policy", "default-src 'self'")
	w.Header().Set("Strict-Transport-Security",
		fmt.Sprintf("max-age=%d; includeSubDomains", hstsMaxAge))
}
