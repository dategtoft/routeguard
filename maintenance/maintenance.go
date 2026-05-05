// Package maintenance provides HTTP middleware that returns a 503 Service
// Unavailable response when the application is in maintenance mode.
package maintenance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
)

// Options configures the maintenance middleware.
type Options struct {
	// Message is the message returned to clients.
	Message string
	// RetryAfter sets the Retry-After header value in seconds. 0 disables it.
	RetryAfter int
	// JSONResponse controls whether the response body is JSON-encoded.
	JSONResponse bool
}

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Message:      "Service temporarily unavailable. Please try again later.",
		RetryAfter:   120,
		JSONResponse: true,
	}
}

type errorBody struct {
	Error string `json:"error"`
}

// Middleware holds the maintenance state and http.Handler wrapper.
type Middleware struct {
	active  atomic.Bool
	options Options
}

// New creates a new Middleware. When active is true requests are rejected
// immediately with 503.
func New(active bool, opts Options) *Middleware {
	m := &Middleware{options: opts}
	m.active.Store(active)
	return m
}

// Enable puts the service into maintenance mode.
func (m *Middleware) Enable() { m.active.Store(true) }

// Disable takes the service out of maintenance mode.
func (m *Middleware) Disable() { m.active.Store(false) }

// Active reports whether maintenance mode is currently enabled.
func (m *Middleware) Active() bool { return m.active.Load() }

// Handler returns an http.Handler that wraps next.
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.active.Load() {
			next.ServeHTTP(w, r)
			return
		}

		if m.options.RetryAfter > 0 {
			w.Header().Set("Retry-After", fmt.Sprintf("%d", m.options.RetryAfter))
		}

		if m.options.JSONResponse {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(errorBody{Error: m.options.Message})
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(m.options.Message))
	})
}
