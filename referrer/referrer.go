package referrer

import (
	"net/http"
	"strings"
)

// Options configures the referrer policy middleware.
type Options struct {
	// Policy sets the Referrer-Policy header value.
	// Defaults to "strict-origin-when-cross-origin".
	Policy string

	// AllowedHosts is an optional list of allowed referrer hostnames.
	// If non-empty, requests with a Referer header not matching any
	// entry will receive a 403 response.
	AllowedHosts []string

	// BlockEmpty controls whether requests with no Referer header are
	// rejected when AllowedHosts is set. Defaults to false.
	BlockEmpty bool
}

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Policy:     "strict-origin-when-cross-origin",
		BlockEmpty: false,
	}
}

// New returns middleware that sets the Referrer-Policy response header and
// optionally validates the incoming Referer request header against an
// allowlist of trusted hostnames.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.Policy == "" {
		opts.Policy = DefaultOptions().Policy
	}

	allowed := make(map[string]struct{}, len(opts.AllowedHosts))
	for _, h := range opts.AllowedHosts {
		allowed[strings.ToLower(h)] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(allowed) > 0 {
				referer := r.Referer()
				if referer == "" {
					if opts.BlockEmpty {
						http.Error(w, "Forbidden", http.StatusForbidden)
						return
					}
				} else {
					host := extractHost(referer)
					if _, ok := allowed[strings.ToLower(host)]; !ok {
						http.Error(w, "Forbidden", http.StatusForbidden)
						return
					}
				}
			}

			w.Header().Set("Referrer-Policy", opts.Policy)
			next.ServeHTTP(w, r)
		})
	}
}

// extractHost parses the hostname from a raw URL string without importing
// net/url to keep allocations minimal for the happy path.
func extractHost(rawURL string) string {
	s := rawURL
	if i := strings.Index(s, "://"); i != -1 {
		s = s[i+3:]
	}
	if i := strings.IndexAny(s, "/:?#"); i != -1 {
		s = s[:i]
	}
	return s
}
