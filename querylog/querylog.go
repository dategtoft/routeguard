// Package querylog provides middleware that logs query parameters from incoming requests.
package querylog

import (
	"log"
	"net/http"
	"os"
	"strings"
)

// Options configures the querylog middleware.
type Options struct {
	// Logger is the logger to write to. Defaults to stderr.
	Logger *log.Logger
	// SkipPaths is a list of paths to skip logging for.
	SkipPaths []string
	// RedactKeys is a list of query parameter keys whose values will be redacted.
	RedactKeys []string
	// Prefix is the log line prefix.
	Prefix string
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Logger: log.New(os.Stderr, "", log.LstdFlags),
		Prefix: "[querylog]",
	}
}

// New returns middleware that logs query parameters for each request.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.Logger == nil {
		opts.Logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	skip := make(map[string]struct{}, len(opts.SkipPaths))
	for _, p := range opts.SkipPaths {
		skip[p] = struct{}{}
	}

	redact := make(map[string]struct{}, len(opts.RedactKeys))
	for _, k := range opts.RedactKeys {
		redact[strings.ToLower(k)] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, skip := skip[r.URL.Path]; !skip {
				q := r.URL.Query()
				if len(q) > 0 {
					parts := make([]string, 0, len(q))
					for k, vals := range q {
						if _, isRedacted := redact[strings.ToLower(k)]; isRedacted {
							parts = append(parts, k+"=[REDACTED]")
						} else {
							parts = append(parts, k+"="+strings.Join(vals, ","))
						}
					}
					opts.Logger.Printf("%s %s %s %s", opts.Prefix, r.Method, r.URL.Path, strings.Join(parts, " "))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
