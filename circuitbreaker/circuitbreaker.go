// Package circuitbreaker provides an HTTP middleware that implements the
// circuit breaker pattern to prevent cascading failures.
package circuitbreaker

import (
	"net/http"
	"sync"
	"time"
)

// State represents the circuit breaker state.
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// Options configures the circuit breaker middleware.
type Options struct {
	// Threshold is the number of consecutive failures before opening the circuit.
	Threshold int
	// Timeout is how long the circuit stays open before moving to half-open.
	Timeout time.Duration
	// StatusCode is the HTTP status code returned when the circuit is open.
	StatusCode int
	// Message is the response body when the circuit is open.
	Message string
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Threshold:  5,
		Timeout:    30 * time.Second,
		StatusCode: http.StatusServiceUnavailable,
		Message:    "service unavailable",
	}
}

type breaker struct {
	mu          sync.Mutex
	state       State
	failures    int
	lastFailure time.Time
	opts        Options
}

// New returns an HTTP middleware that wraps the next handler with circuit
// breaker logic based on 5xx response codes.
func New(opts Options) func(http.Handler) http.Handler {
	b := &breaker{opts: opts}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !b.allow() {
				http.Error(w, opts.Message, opts.StatusCode)
				return
			}

			rec := &statusRecorder{ResponseWriter: w, code: http.StatusOK}
			next.ServeHTTP(rec, r)

			if rec.code >= 500 {
				b.recordFailure()
			} else {
				b.recordSuccess()
			}
		})
	}
}

func (b *breaker) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(b.lastFailure) >= b.opts.Timeout {
			b.state = StateHalfOpen
			return true
		}
		return false
	case StateHalfOpen:
		return true
	}
	return false
}

func (b *breaker) recordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures++
	b.lastFailure = time.Now()
	if b.failures >= b.opts.Threshold {
		b.state = StateOpen
	}
}

func (b *breaker) recordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures = 0
	b.state = StateClosed
}

type statusRecorder struct {
	http.ResponseWriter
	code int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.code = code
	s.ResponseWriter.WriteHeader(code)
}
