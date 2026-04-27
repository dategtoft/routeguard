package requestid

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const (
	// DefaultHeader is the default HTTP header used to propagate the request ID.
	DefaultHeader = "X-Request-ID"

	// contextKey is the key used to store the request ID in the context.
	contextKey contextKeyType = iota
)

type contextKeyType int

// Options holds configuration for the request ID middleware.
type Options struct {
	// Header is the HTTP header name to read/write the request ID.
	Header string

	// Generator is a custom function for generating request IDs.
	// Defaults to a 16-byte random hex string.
	Generator func() string
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Header:    DefaultHeader,
		Generator: defaultGenerator,
	}
}

// New returns middleware that attaches a unique request ID to every request.
// If the incoming request already carries the configured header, that value is
// reused; otherwise a new ID is generated. The ID is stored in the request
// context and echoed back in the response header.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.Header == "" {
		opts.Header = DefaultHeader
	}
	if opts.Generator == nil {
		opts.Generator = defaultGenerator
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get(opts.Header)
			if id == "" {
				id = opts.Generator()
			}

			ctx := context.WithValue(r.Context(), contextKey, id)
			w.Header().Set(opts.Header, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// FromContext retrieves the request ID stored in the context.
// Returns an empty string if no ID is present.
func FromContext(ctx context.Context) string {
	if id, ok := ctx.Value(contextKey).(string); ok {
		return id
	}
	return ""
}

func defaultGenerator() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
