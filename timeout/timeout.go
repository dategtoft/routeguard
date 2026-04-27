package timeout

import (
	"context"
	"net/http"
	"time"
)

// Options holds configuration for the timeout middleware.
type Options struct {
	// Duration is the maximum time allowed for a request to complete.
	Duration time.Duration
	// Message is the response body sent when a timeout occurs.
	Message string
	// StatusCode is the HTTP status code sent on timeout.
	StatusCode int
}

// DefaultOptions returns sensible defaults for the timeout middleware.
func DefaultOptions() Options {
	return Options{
		Duration:   30 * time.Second,
		Message:    "request timed out",
		StatusCode: http.StatusGatewayTimeout,
	}
}

// New returns a middleware that cancels requests exceeding the configured duration.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.Duration <= 0 {
		opts.Duration = DefaultOptions().Duration
	}
	if opts.Message == "" {
		opts.Message = DefaultOptions().Message
	}
	if opts.StatusCode == 0 {
		opts.StatusCode = DefaultOptions().StatusCode
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), opts.Duration)
			defer cancel()

			done := make(chan struct{})
			tw := &timeoutWriter{ResponseWriter: w}

			go func() {
				next.ServeHTTP(tw, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				tw.flush(w)
			case <-ctx.Done():
				w.WriteHeader(opts.StatusCode)
				w.Write([]byte(opts.Message)) //nolint:errcheck
			}
		})
	}
}

// timeoutWriter buffers the response so we can discard it on timeout.
type timeoutWriter struct {
	http.ResponseWriter
	code int
	body []byte
	wroteHeader bool
}

func (tw *timeoutWriter) WriteHeader(code int) {
	if !tw.wroteHeader {
		tw.code = code
		tw.wroteHeader = true
	}
}

func (tw *timeoutWriter) Write(b []byte) (int, error) {
	tw.body = append(tw.body, b...)
	return len(b), nil
}

func (tw *timeoutWriter) flush(w http.ResponseWriter) {
	if tw.wroteHeader {
		w.WriteHeader(tw.code)
	}
	if len(tw.body) > 0 {
		w.Write(tw.body) //nolint:errcheck
	}
}
