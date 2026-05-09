// Package mimetype provides middleware that validates and enforces
// response Content-Type headers set by downstream handlers.
package mimetype

import (
	"net/http"
	"strings"
)

// Options configures the mimetype middleware.
type Options struct {
	// Allowed is the list of permitted MIME types for responses.
	// Wildcards like "image/*" are supported.
	Allowed []string

	// OnReject is called when the response Content-Type is not allowed.
	// Defaults to a 406 Not Acceptable response.
	OnReject func(w http.ResponseWriter, r *http.Request)
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Allowed: []string{"application/json", "text/plain", "text/html"},
	}
}

type captureWriter struct {
	http.ResponseWriter
	code int
	sniffed string
	wroteHeader bool
}

func (c *captureWriter) WriteHeader(code int) {
	c.code = code
	c.wroteHeader = true
}

func (c *captureWriter) Header() http.Header {
	return c.ResponseWriter.Header()
}

// New returns middleware that rejects responses whose Content-Type does not
// match any entry in Options.Allowed.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.OnReject == nil {
		opts.OnReject = func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "406 Not Acceptable", http.StatusNotAcceptable)
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cw := &captureWriter{ResponseWriter: w, code: http.StatusOK}
			next.ServeHTTP(cw, r)

			ct := w.Header().Get("Content-Type")
			if ct == "" || isAllowed(ct, opts.Allowed) {
				if cw.wroteHeader {
					w.WriteHeader(cw.code)
				}
				return
			}

			w.Header().Del("Content-Type")
			opts.OnReject(w, r)
		})
	}
}

// isAllowed reports whether ct matches any pattern in allowed.
func isAllowed(ct string, allowed []string) bool {
	mime := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
	for _, a := range allowed {
		a = strings.ToLower(strings.TrimSpace(a))
		if a == mime {
			return true
		}
		if strings.HasSuffix(a, "/*") {
			prefix := strings.TrimSuffix(a, "*")
			if strings.HasPrefix(mime, prefix) {
				return true
			}
		}
	}
	return false
}
