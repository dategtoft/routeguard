// Package middleware composes routeguard middleware components into a single chain.
package middleware

import (
	"net/http"

	"github.com/yourusername/routeguard/cors"
	"github.com/yourusername/routeguard/jwt"
	"github.com/yourusername/routeguard/logger"
	"github.com/yourusername/routeguard/ratelimit"
	"github.com/yourusername/routeguard/recovery"
)

// Options configures which middleware components are enabled.
type Options struct {
	RateLimit *ratelimit.Options
	JWT       *jwt.Options
	CORS      *cors.Options
	Logger    *logger.Options
	Recovery  *recovery.Options
}

// New builds and returns a composed middleware chain wrapping the given handler.
// Middleware is applied in the following order (outermost first):
// Recovery -> Logger -> CORS -> RateLimit -> JWT -> handler
func New(next http.Handler, opts Options) http.Handler {
	h := next

	if opts.JWT != nil {
		h = jwt.New(h, *opts.JWT)
	}

	if opts.RateLimit != nil {
		h = ratelimit.New(h, *opts.RateLimit)
	}

	if opts.CORS != nil {
		h = cors.New(h, *opts.CORS)
	}

	if opts.Logger != nil {
		h = logger.New(h, *opts.Logger)
	}

	if opts.Recovery != nil {
		h = recovery.New(h, *opts.Recovery)
	}

	return h
}
