package middleware

import (
	"net/http"

	"github.com/yourusername/routeguard/jwt"
	"github.com/yourusername/routeguard/logger"
	"github.com/yourusername/routeguard/ratelimit"
)

// Options holds optional middleware components.
type Options struct {
	RateLimiter *ratelimit.RateLimiter
	JWT         *jwt.JWT
	Logger      *logger.Logger
}

// Chain wraps a handler with the configured middleware stack.
type Chain struct {
	opts Options
}

// New creates a new middleware Chain with the provided options.
func New(opts Options) *Chain {
	return &Chain{opts: opts}
}

// Handler wraps the given http.Handler with all enabled middleware.
func (c *Chain) Handler(next http.Handler) http.Handler {
	handler := next

	if c.opts.RateLimiter != nil {
		handler = c.opts.RateLimiter.Middleware(handler)
	}

	if c.opts.JWT != nil {
		handler = c.opts.JWT.Middleware(handler)
	}

	if c.opts.Logger != nil {
		handler = c.opts.Logger.Middleware(handler)
	}

	return handler
}

// HandlerFunc wraps the given http.HandlerFunc with all enabled middleware.
func (c *Chain) HandlerFunc(fn http.HandlerFunc) http.Handler {
	return c.Handler(fn)
}
