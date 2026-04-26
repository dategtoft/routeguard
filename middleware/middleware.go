// Package middleware provides a unified interface for combining routeguard's
// rate limiting and JWT validation into a single composable middleware chain.
package middleware

import (
	"net/http"

	"github.com/yourusername/routeguard/jwt"
	"github.com/yourusername/routeguard/ratelimit"
)

// Options holds configuration for the combined middleware.
type Options struct {
	// RateLimit configures the rate limiter. If nil, rate limiting is disabled.
	RateLimit *ratelimit.Options
	// JWT configures the JWT validator. If nil, JWT validation is disabled.
	JWT *jwt.Options
}

// RateLimitOptions is an alias for configuring rate limiting within middleware.
type RateLimitOptions = ratelimit.Options

// JWTOptions is an alias for configuring JWT validation within middleware.
type JWTOptions = jwt.Options

// Chain holds the composed middleware handlers.
type Chain struct {
	rl  *ratelimit.Limiter
	jwt *jwt.Validator
}

// New creates a new middleware Chain based on the provided Options.
// Only non-nil option blocks will be activated.
func New(opts Options) *Chain {
	c := &Chain{}

	if opts.RateLimit != nil {
		c.rl = ratelimit.New(*opts.RateLimit)
	}

	if opts.JWT != nil {
		c.jwt = jwt.New(*opts.JWT)
	}

	return c
}

// Handler wraps the given http.Handler with all configured middleware.
// Middleware is applied in the following order:
//  1. Rate limiting (if configured)
//  2. JWT validation (if configured)
//
// If a middleware rejects the request, subsequent middleware will not run.
func (c *Chain) Handler(next http.Handler) http.Handler {
	h := next

	// Apply JWT middleware first in the wrap so it runs after rate limiting.
	if c.jwt != nil {
		h = c.jwt.Middleware(h)
	}

	// Rate limiting is the outermost layer — checked before JWT validation.
	if c.rl != nil {
		h = c.rl.Middleware(h)
	}

	return h
}

// HandlerFunc wraps the given http.HandlerFunc with all configured middleware.
func (c *Chain) HandlerFunc(fn http.HandlerFunc) http.Handler {
	return c.Handler(fn)
}
