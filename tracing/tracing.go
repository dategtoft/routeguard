// Package tracing provides HTTP middleware that injects and propagates
// trace IDs across requests using configurable headers.
package tracing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// contextKey is an unexported type for context keys in this package.
type contextKey struct{}

// Options configures the tracing middleware.
type Options struct {
	// TraceHeader is the HTTP header used to read/write the trace ID.
	// Defaults to "X-Trace-Id".
	TraceHeader string

	// SpanHeader is the HTTP header used to write the span ID.
	// Defaults to "X-Span-Id".
	SpanHeader string

	// Generator is an optional custom function to generate trace IDs.
	// Defaults to a 16-byte random hex string.
	Generator func() string
}

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		TraceHeader: "X-Trace-Id",
		SpanHeader:  "X-Span-Id",
		Generator:   defaultGenerator,
	}
}

// FromContext returns the trace ID stored in ctx, or an empty string.
func FromContext(ctx context.Context) string {
	v, _ := ctx.Value(contextKey{}).(string)
	return v
}

// New returns middleware that ensures every request carries a trace ID.
// If the incoming request already has a trace ID header it is reused;
// otherwise a new one is generated. A new span ID is always generated.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.TraceHeader == "" {
		opts.TraceHeader = "X-Trace-Id"
	}
	if opts.SpanHeader == "" {
		opts.SpanHeader = "X-Span-Id"
	}
	if opts.Generator == nil {
		opts.Generator = defaultGenerator
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := r.Header.Get(opts.TraceHeader)
			if traceID == "" {
				traceID = opts.Generator()
			}
			spanID := opts.Generator()

			w.Header().Set(opts.TraceHeader, traceID)
			w.Header().Set(opts.SpanHeader, spanID)

			ctx := context.WithValue(r.Context(), contextKey{}, traceID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func defaultGenerator() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
