package cors

import (
	"net/http"
	"strings"
)

// Options holds the configuration for CORS middleware.
type Options struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	AllowCredentials bool
}

// DefaultOptions returns a permissive default CORS configuration.
func DefaultOptions() Options {
	return Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
		AllowCredentials: false,
	}
}

type cors struct {
	opts Options
}

// New returns a new CORS middleware handler.
func New(opts Options) func(http.Handler) http.Handler {
	c := &cors{opts: opts}
	return c.middleware
}

func (c *cors) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && c.isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if len(c.opts.AllowedOrigins) == 1 && c.opts.AllowedOrigins[0] == "*" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", strings.Join(c.opts.AllowedMethods, ", "))
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(c.opts.AllowedHeaders, ", "))

		if c.opts.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (c *cors) isOriginAllowed(origin string) bool {
	for _, allowed := range c.opts.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}
