// Package shield provides a middleware that sets common security-related
// HTTP response headers to help protect against well-known web vulnerabilities.
package shield

import "net/http"

// Options configures the Shield middleware.
type Options struct {
	// ContentSecurityPolicy sets the Content-Security-Policy header.
	// Defaults to "default-src 'self'".
	ContentSecurityPolicy string

	// XContentTypeOptions sets the X-Content-Type-Options header.
	// Defaults to "nosniff".
	XContentTypeOptions string

	// XFrameOptions sets the X-Frame-Options header.
	// Defaults to "DENY".
	XFrameOptions string

	// XXSSProtection sets the X-XSS-Protection header.
	// Defaults to "1; mode=block".
	XXSSProtection string

	// ReferrerPolicy sets the Referrer-Policy header.
	// Defaults to "strict-origin-when-cross-origin".
	ReferrerPolicy string

	// PermissionsPolicy sets the Permissions-Policy header.
	// Empty by default (header omitted).
	PermissionsPolicy string
}

// DefaultOptions returns an Options with sensible security defaults.
func DefaultOptions() Options {
	return Options{
		ContentSecurityPolicy: "default-src 'self'",
		XContentTypeOptions:   "nosniff",
		XFrameOptions:         "DENY",
		XXSSProtection:        "1; mode=block",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}
}

// New returns a middleware that injects security headers into every response.
func New(opts Options) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			if opts.ContentSecurityPolicy != "" {
				h.Set("Content-Security-Policy", opts.ContentSecurityPolicy)
			}
			if opts.XContentTypeOptions != "" {
				h.Set("X-Content-Type-Options", opts.XContentTypeOptions)
			}
			if opts.XFrameOptions != "" {
				h.Set("X-Frame-Options", opts.XFrameOptions)
			}
			if opts.XXSSProtection != "" {
				h.Set("X-XSS-Protection", opts.XXSSProtection)
			}
			if opts.ReferrerPolicy != "" {
				h.Set("Referrer-Policy", opts.ReferrerPolicy)
			}
			if opts.PermissionsPolicy != "" {
				h.Set("Permissions-Policy", opts.PermissionsPolicy)
			}
			next.ServeHTTP(w, r)
		})
	}
}
