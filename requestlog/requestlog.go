// Package requestlog provides structured HTTP request logging middleware.
package requestlog

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Options configures the request log middleware.
type Options struct {
	// Writer is the output destination. Defaults to os.Stdout.
	Writer io.Writer
	// TimeFormat is the format used for timestamps. Defaults to time.RFC3339.
	TimeFormat string
	// SkipPaths is a list of URL paths that will not be logged.
	SkipPaths []string
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Writer:     os.Stdout,
		TimeFormat: time.RFC3339,
	}
}

type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

// New returns a middleware that logs each request in a structured single-line
// format: timestamp method path status bytes duration.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}
	if opts.TimeFormat == "" {
		opts.TimeFormat = time.RFC3339
	}
	skip := make(map[string]struct{}, len(opts.SkipPaths))
	for _, p := range opts.SkipPaths {
		skip[p] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := skip[r.URL.Path]; ok {
				next.ServeHTTP(w, r)
				return
			}
			start := time.Now()
			rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			duration := time.Since(start)
			fmt.Fprintf(opts.Writer, "time=%s method=%s path=%s status=%d bytes=%d duration=%s\n",
				start.Format(opts.TimeFormat),
				r.Method,
				r.URL.Path,
				rec.status,
				rec.bytes,
				duration,
			)
		})
	}
}
