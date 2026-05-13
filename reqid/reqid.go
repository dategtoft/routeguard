// Package reqid provides middleware that enforces a required request ID on
// incoming HTTP requests, rejecting any request that lacks one.
package reqid

import (
	"net/http"
)

// Options configures the required-request-ID middleware.
type Options struct {
	// Header is the HTTP header name to inspect. Defaults to "X-Request-ID".
	Header string
	// Message is the response body sent when the header is missing.
	Message string
	// StatusCode is the HTTP status returned when the header is absent.
	// Defaults to 400 Bad Request.
	StatusCode int
}

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Header:     "X-Request-ID",
		Message:    "missing required request ID",
		StatusCode: http.StatusBadRequest,
	}
}

// New returns middleware that rejects requests missing the configured header.
// If the header is present and non-empty the request is passed to next.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.Header == "" {
		opts.Header = DefaultOptions().Header
	}
	if opts.Message == "" {
		opts.Message = DefaultOptions().Message
	}
	if opts.StatusCode == 0 {
		opts.StatusCode = DefaultOptions().StatusCode
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get(opts.Header) == "" {
				http.Error(w, opts.Message, opts.StatusCode)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
