// Package depthcheck provides middleware that limits JSON request body nesting depth.
package depthcheck

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Options configures the depthcheck middleware.
type Options struct {
	// MaxDepth is the maximum allowed JSON nesting depth. Default: 10.
	MaxDepth int
	// Message is the response body sent when depth is exceeded.
	Message string
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		MaxDepth: 10,
		Message:  "request body nesting depth exceeded",
	}
}

// New returns middleware that rejects requests whose JSON body exceeds MaxDepth.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.MaxDepth <= 0 {
		opts.MaxDepth = DefaultOptions().MaxDepth
	}
	if opts.Message == "" {
		opts.Message = DefaultOptions().Message
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ct := r.Header.Get("Content-Type")
			if r.Body == nil || !strings.Contains(ct, "application/json") {
				next.ServeHTTP(w, r)
				return
			}

			dec := json.NewDecoder(r.Body)
			defer r.Body.Close()

			if exceeded := scanDepth(dec, opts.MaxDepth); exceeded {
				http.Error(w, opts.Message, http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// scanDepth reads JSON tokens and tracks nesting depth.
func scanDepth(dec *json.Decoder, max int) bool {
	depth := 0
	for {
		t, err := dec.Token()
		if err != nil {
			break
		}
		switch t {
		case json.Delim('{'), json.Delim('['):
			depth++
			if depth > max {
				return true
			}
		case json.Delim('}'), json.Delim(']'):
			depth--
		}
	}
	return false
}
