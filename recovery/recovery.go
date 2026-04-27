// Package recovery provides HTTP middleware that recovers from panics
// and returns a 500 Internal Server Error response.
package recovery

import (
	"log"
	"net/http"
	"runtime/debug"
)

// Options configures the recovery middleware.
type Options struct {
	// Logger is used to log panic details. Defaults to the standard logger.
	Logger *log.Logger
	// EnableStackTrace controls whether the stack trace is logged.
	EnableStackTrace bool
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Logger:           log.Default(),
		EnableStackTrace: true,
	}
}

type middleware struct {
	next http.Handler
	opts Options
}

// New wraps the given handler with panic recovery middleware.
func New(next http.Handler, opts Options) http.Handler {
	if opts.Logger == nil {
		opts.Logger = log.Default()
	}
	return &middleware{next: next, opts: opts}
}

func (m *middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			if m.opts.EnableStackTrace {
				m.opts.Logger.Printf("[recovery] panic: %v\n%s", rec, debug.Stack())
			} else {
				m.opts.Logger.Printf("[recovery] panic: %v", rec)
			}
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}()
	m.next.ServeHTTP(w, r)
}
