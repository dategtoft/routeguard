package middleware

import (
	"net/http"

	"github.com/yourusername/routeguard/cors"
	"github.com/yourusername/routeguard/jwt"
	"github.com/yourusername/routeguard/logger"
	"github.com/yourusername/routeguard/ratelimit"
)

// Options configures the combined middleware stack.
type Options struct {
	RateLimit *ratelimit.Options
	JWT       *jwt.Options
	Logger    *logger.Options
	CORS      *cors.Options
}

// New builds and returns a composed middleware chain based on the provided Options.
// Middlewares are applied in order: Logger → CORS → RateLimit → JWT.
func New(opts Options) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		h := next

		if opts.JWT != nil {
			h = jwt.New(*opts.JWT)(h)
		}

		if opts.RateLimit != nil {
			h = ratelimit.New(*opts.RateLimit)(h)
		}

		if opts.CORS != nil {
			h = cors.New(*opts.CORS)(h)
		}

		if opts.Logger != nil {
			h = logger.New(*opts.Logger)(h)
		}

		return h
	}
}
