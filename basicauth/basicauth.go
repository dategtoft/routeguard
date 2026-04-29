package basicauth

import (
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"
)

// Options configures the Basic Auth middleware.
type Options struct {
	// Realm is the authentication realm shown in the browser prompt.
	Realm string
	// Credentials is a map of username to password.
	Credentials map[string]string
	// UnauthorizedHandler is called when authentication fails.
	// If nil, a default 401 response is sent.
	UnauthorizedHandler http.HandlerFunc
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Realm:       "Restricted",
		Credentials: make(map[string]string),
	}
}

// New returns a Basic Auth middleware using the provided options.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.Realm == "" {
		opts.Realm = "Restricted"
	}
	if opts.Credentials == nil {
		opts.Credentials = make(map[string]string)
	}
	unauthorized := opts.UnauthorizedHandler
	if unauthorized == nil {
		unauthorized = func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+opts.Realm+`"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := parseBasicAuth(r)
			if !ok {
				unauthorized(w, r)
				return
			}
			expected, exists := opts.Credentials[user]
			if !exists || !secureCompare(expected, pass) {
				unauthorized(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// parseBasicAuth extracts username and password from the Authorization header.
func parseBasicAuth(r *http.Request) (string, string, bool) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Basic ") {
		return "", "", false
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, "Basic "))
	if err != nil {
		return "", "", false
	}
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

// secureCompare performs a constant-time string comparison.
func secureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
