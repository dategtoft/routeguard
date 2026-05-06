// Package paginate provides HTTP middleware that parses and validates
// pagination query parameters (page, limit) and injects them into the
// request context for downstream handlers.
package paginate

import (
	"context"
	"net/http"
	"strconv"
)

type contextKey struct{}

// Params holds the parsed pagination values.
type Params struct {
	Page  int
	Limit int
	Offset int
}

// Options configures the pagination middleware.
type Options struct {
	// DefaultLimit is used when the client omits the limit parameter.
	DefaultLimit int
	// MaxLimit caps the limit value to prevent excessive queries.
	MaxLimit int
	// PageParam is the query parameter name for the page number (default: "page").
	PageParam string
	// LimitParam is the query parameter name for the page size (default: "limit").
	LimitParam string
}

// DefaultOptions returns sensible defaults for the pagination middleware.
func DefaultOptions() Options {
	return Options{
		DefaultLimit: 20,
		MaxLimit:     100,
		PageParam:    "page",
		LimitParam:   "limit",
	}
}

// FromContext retrieves Params from the request context.
// Returns zero-value Params if none are present.
func FromContext(ctx context.Context) Params {
	v, _ := ctx.Value(contextKey{}).(Params)
	return v
}

// New returns middleware that parses pagination query parameters and stores
// the result in the request context. Invalid or out-of-range values cause a
// 400 Bad Request response.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.DefaultLimit <= 0 {
		opts.DefaultLimit = DefaultOptions().DefaultLimit
	}
	if opts.MaxLimit <= 0 {
		opts.MaxLimit = DefaultOptions().MaxLimit
	}
	if opts.PageParam == "" {
		opts.PageParam = DefaultOptions().PageParam
	}
	if opts.LimitParam == "" {
		opts.LimitParam = DefaultOptions().LimitParam
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()

			page := 1
			if raw := q.Get(opts.PageParam); raw != "" {
				v, err := strconv.Atoi(raw)
				if err != nil || v < 1 {
					http.Error(w, "invalid page parameter", http.StatusBadRequest)
					return
				}
				page = v
			}

			limit := opts.DefaultLimit
			if raw := q.Get(opts.LimitParam); raw != "" {
				v, err := strconv.Atoi(raw)
				if err != nil || v < 1 {
					http.Error(w, "invalid limit parameter", http.StatusBadRequest)
					return
				}
				if v > opts.MaxLimit {
					v = opts.MaxLimit
				}
				limit = v
			}

			params := Params{
				Page:   page,
				Limit:  limit,
				Offset: (page - 1) * limit,
			}
			ctx := context.WithValue(r.Context(), contextKey{}, params)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
