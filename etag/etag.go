// Package etag provides HTTP ETag middleware for conditional request handling.
package etag

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
)

// Options configures the ETag middleware.
type Options struct {
	// Weak indicates whether to generate weak ETags (W/"...").
	Weak bool
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Weak: false,
	}
}

// responseBuffer captures the response body and status for ETag generation.
type responseBuffer struct {
	http.ResponseWriter
	buf    bytes.Buffer
	status int
}

func (rb *responseBuffer) WriteHeader(status int) {
	rb.status = status
}

func (rb *responseBuffer) Write(b []byte) (int, error) {
	return rb.buf.Write(b)
}

// New returns ETag middleware that generates and validates ETags for responses.
// If the client sends a matching If-None-Match header, a 304 Not Modified is returned.
func New(opts Options) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only handle GET and HEAD requests.
			if r.Method != http.MethodGet && r.Method != http.MethodHead {
				next.ServeHTTP(w, r)
				return
			}

			rb := &responseBuffer{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rb, r)

			body := rb.buf.Bytes()
			etag := generateETag(body, opts.Weak)

			w.Header().Set("ETag", etag)

			if match := r.Header.Get("If-None-Match"); match != "" {
				if matchesETag(match, etag) {
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}

			w.WriteHeader(rb.status)
			if r.Method != http.MethodHead {
				w.Write(body) //nolint:errcheck
			}
		})
	}
}

func generateETag(body []byte, weak bool) string {
	sum := sha256.Sum256(body)
	hash := fmt.Sprintf("%x", sum[:8])
	if weak {
		return `W/"` + hash + `"`
	}
	return `"` + hash + `"`
}

func matchesETag(header, etag string) bool {
	for _, part := range strings.Split(header, ",") {
		if strings.TrimSpace(part) == etag {
			return true
		}
	}
	return false
}
