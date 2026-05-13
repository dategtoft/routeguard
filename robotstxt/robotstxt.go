// Package robotstxt provides middleware that serves a robots.txt response
// and optionally blocks requests from known bot user-agents.
package robotstxt

import (
	"net/http"
	"strings"
)

// Options configures the robotstxt middleware.
type Options struct {
	// Content is the body returned at /robots.txt. Defaults to "User-agent: *\nDisallow:"
	Content string
	// BlockBots, when true, returns 403 for requests whose User-Agent matches
	// a known crawler/bot pattern.
	BlockBots bool
	// BlockedAgents is a list of sub-strings (case-insensitive) matched against
	// the User-Agent header when BlockBots is enabled.
	BlockedAgents []string
	// Path is the URL path that serves robots.txt. Defaults to "/robots.txt".
	Path string
}

// DefaultOptions returns a sensible default configuration.
func DefaultOptions() Options {
	return Options{
		Content:   "User-agent: *\nDisallow:",
		BlockBots: false,
		BlockedAgents: []string{
			"googlebot", "bingbot", "slurp", "duckduckbot",
			"baiduspider", "yandexbot", "sogou", "exabot",
		},
		Path: "/robots.txt",
	}
}

// New returns middleware that serves robots.txt and optionally blocks bots.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.Path == "" {
		opts.Path = "/robots.txt"
	}
	if opts.Content == "" {
		opts.Content = "User-agent: *\nDisallow:"
	}
	agents := make([]string, len(opts.BlockedAgents))
	for i, a := range opts.BlockedAgents {
		agents[i] = strings.ToLower(a)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Serve robots.txt at the configured path.
			if r.URL.Path == opts.Path {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(opts.Content))
				return
			}

			// Optionally block known bots.
			if opts.BlockBots {
				ua := strings.ToLower(r.Header.Get("User-Agent"))
				for _, bot := range agents {
					if strings.Contains(ua, bot) {
						http.Error(w, "Forbidden", http.StatusForbidden)
						return
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
