package middleware

import (
	"net/http"

	"github.com/yourusername/routeguard/cors"
	"github.com/yourusername/routeguard/jwt"
	"github.com/yourusername/routeguard/logger"
	"github.com/yourusername/routeguard/ratelimit"
	"github.com/yourusername/routeguard/recovery"
	"github.com/yourusername/routeguard/timeout"
)

// Options configures which middleware components are enabled.
type Options struct {
	RateLimit *ratelimit.Options
	JWT       *jwt.Options
	Logger    *logger.Options
	CORS      *cors.Options
	Recovery  *recovery.Options
	Timeout   *timeout.Options
}

// New builds a middleware chain based on the provided options.
// Middleware is applied in the following order (outermost first):
// Recovery → Logger → CORS → Timeout → RateLimit → JWT → handler
func New(opts Options) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		h := next

		if opts.JWT != nil {
			h = jwt.New(*opts.JWT)(h)
		}

		if opts.RateLimit != nil {
			h = ratelimit.New(*opts.RateLimit)(h)
		}

		if opts.Timeout != nil {
			h = timeout.New(*opts.Timeout)(h)
		}

		if opts.CORS != nil {
			h = cors.New(*opts.CORS)(h)
		}

		if opts.Logger != nil {
			h = logger.New(*opts.Logger)(h)
		}

		if opts.Recovery != nil {
			h = recovery.New(*opts.Recovery)(h)
		}

		return h
	}
}

// Chain applies a slice of middleware functions to a handler in order,
// so that the first middleware in the slice is the outermost wrapper.
// This allows composing middleware independently of the Options struct.
func Chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}
