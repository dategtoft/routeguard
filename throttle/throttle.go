// Package throttle provides a concurrency-limiting middleware that caps the
// number of requests handled simultaneously, queuing or rejecting excess ones.
package throttle

import (
	"net/http"
	"time"
)

// Options configures the throttle middleware.
type Options struct {
	// MaxConcurrent is the maximum number of requests processed at once.
	MaxConcurrent int
	// MaxQueueSize is the number of requests allowed to wait. 0 means no queue.
	MaxQueueSize int
	// Timeout is how long a queued request waits before being rejected.
	Timeout time.Duration
	// StatusCode is returned when the limit is exceeded (default 503).
	StatusCode int
	// Message is the response body when the limit is exceeded.
	Message string
}

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		MaxConcurrent: 100,
		MaxQueueSize:  50,
		Timeout:       5 * time.Second,
		StatusCode:    http.StatusServiceUnavailable,
		Message:       "Too many concurrent requests",
	}
}

type middleware struct {
	opts Options
	sem  chan struct{}
	queue chan struct{}
}

// New returns an HTTP middleware that limits concurrent request processing.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.MaxConcurrent <= 0 {
		opts.MaxConcurrent = DefaultOptions().MaxConcurrent
	}
	if opts.StatusCode == 0 {
		opts.StatusCode = DefaultOptions().StatusCode
	}
	if opts.Message == "" {
		opts.Message = DefaultOptions().Message
	}

	m := &middleware{
		opts:  opts,
		sem:   make(chan struct{}, opts.MaxConcurrent),
		queue: make(chan struct{}, opts.MaxQueueSize),
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case m.sem <- struct{}{}:
				// Slot acquired immediately.
				defer func() { <-m.sem }()
				next.ServeHTTP(w, r)
			default:
				// No slot; try to queue.
				select {
				case m.queue <- struct{}{}:
				default:
					// Queue full — reject immediately.
					http.Error(w, m.opts.Message, m.opts.StatusCode)
					return
				}
				defer func() { <-m.queue }()

				timer := time.NewTimer(m.opts.Timeout)
				defer timer.Stop()

				select {
				case m.sem <- struct{}{}:
					defer func() { <-m.sem }()
					next.ServeHTTP(w, r)
				case <-timer.C:
					http.Error(w, m.opts.Message, m.opts.StatusCode)
				}
			}
		})
	}
}
