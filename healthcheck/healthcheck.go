// Package healthcheck provides a lightweight health check middleware
// and handler for HTTP services.
package healthcheck

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Status represents the health status of the service.
type Status string

const (
	StatusOK      Status = "ok"
	StatusDegraded Status = "degraded"
)

// Response is the JSON payload returned by the health endpoint.
type Response struct {
	Status    Status            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]string `json:"checks,omitempty"`
}

// CheckFunc is a named health check function that returns an error if unhealthy.
type CheckFunc func() error

// Options configures the health check handler.
type Options struct {
	// Path is the HTTP path for the health endpoint (default: "/healthz").
	Path string
	// Checks is an optional map of named dependency checks.
	Checks map[string]CheckFunc
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Path:   "/healthz",
		Checks: make(map[string]CheckFunc),
	}
}

type handler struct {
	opts Options
	mu   sync.RWMutex
}

// New returns an http.Handler that responds to health check requests.
func New(opts Options) http.Handler {
	if opts.Path == "" {
		opts.Path = "/healthz"
	}
	if opts.Checks == nil {
		opts.Checks = make(map[string]CheckFunc)
	}
	return &handler{opts: opts}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	resp := Response{
		Status:    StatusOK,
		Timestamp: time.Now().UTC(),
		Checks:    make(map[string]string),
	}

	for name, fn := range h.opts.Checks {
		if err := fn(); err != nil {
			resp.Checks[name] = err.Error()
			resp.Status = StatusDegraded
		} else {
			resp.Checks[name] = "ok"
		}
	}

	statusCode := http.StatusOK
	if resp.Status == StatusDegraded {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(resp)
}
