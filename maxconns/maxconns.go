// Package maxconns provides middleware that limits the number of concurrent
// requests being handled by the server at any given time.
package maxconns

import (
	"net/http"
	"sync/atomic"
)

// Options configures the max connections middleware.
type Options struct {
	// Max is the maximum number of concurrent requests allowed.
	// Requests exceeding this limit receive a 503 response.
	Max int64

	// Message is the response body sent when the limit is exceeded.
	Message string
}

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Max:     100,
		Message: "Service Unavailable",
	}
}

type middleware struct {
	opts    Options
	active  atomic.Int64
	next    http.Handler
}

// New returns an http.Handler that limits concurrent requests to opts.Max.
// If opts is zero-value, DefaultOptions is used.
func New(next http.Handler, opts Options) http.Handler {
	if opts.Max <= 0 {
		opts = DefaultOptions()
	}
	if opts.Message == "" {
		opts.Message = DefaultOptions().Message
	}
	return &middleware{opts: opts, next: next}
}

func (m *middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	current := m.active.Add(1)
	defer m.active.Add(-1)

	if current > m.opts.Max {
		http.Error(w, m.opts.Message, http.StatusServiceUnavailable)
		return
	}

	m.next.ServeHTTP(w, r)
}
