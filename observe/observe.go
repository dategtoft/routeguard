// Package observe provides a middleware that exposes a simple metrics
// snapshot (request count, error count, in-flight requests) at a
// configurable path without requiring an external dependency.
package observe

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
)

// Options configures the observe middleware.
type Options struct {
	// Path is the HTTP path that serves the metrics snapshot.
	// Defaults to "/metrics".
	Path string

	// Namespace is an optional string prepended to every metric key.
	Namespace string
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Path: "/metrics",
	}
}

type metrics struct {
	requests  atomic.Int64
	errors    atomic.Int64
	inFlight  atomic.Int64
	startedAt time.Time
}

type middleware struct {
	opts    Options
	stats   *metrics
}

// New returns an http.Handler that wraps next, collecting basic counters and
// serving them as JSON on opts.Path.
func New(next http.Handler, opts Options) http.Handler {
	if opts.Path == "" {
		opts.Path = "/metrics"
	}
	m := &middleware{
		opts:  opts,
		stats: &metrics{startedAt: time.Now()},
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == m.opts.Path {
			m.serveMetrics(w, r)
			return
		}
		m.stats.requests.Add(1)
		m.stats.inFlight.Add(1)
		defer m.stats.inFlight.Add(-1)

		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		if rw.status >= http.StatusInternalServerError {
			m.stats.errors.Add(1)
		}
	})
}

func (m *middleware) serveMetrics(w http.ResponseWriter, _ *http.Request) {
	prefix := m.opts.Namespace
	if prefix != "" {
		prefix += "_"
	}
	payload := map[string]any{
		prefix + "requests_total":   m.stats.requests.Load(),
		prefix + "errors_total":     m.stats.errors.Load(),
		prefix + "in_flight":        m.stats.inFlight.Load(),
		prefix + "uptime_seconds":   time.Since(m.stats.startedAt).Seconds(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(payload)
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
