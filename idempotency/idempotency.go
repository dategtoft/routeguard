// Package idempotency provides middleware that caches responses keyed by an
// idempotency key header so that retried requests with the same key receive
// the original response without re-executing the handler.
package idempotency

import (
	"bytes"
	"net/http"
	"sync"
	"time"
)

// Options configures the idempotency middleware.
type Options struct {
	// Header is the request header that carries the idempotency key.
	// Defaults to "Idempotency-Key".
	Header string

	// TTL is how long a cached response is retained.
	// Defaults to 24 hours.
	TTL time.Duration

	// Methods lists the HTTP methods that are subject to idempotency checks.
	// Defaults to [POST, PATCH].
	Methods []string
}

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Header:  "Idempotency-Key",
		TTL:     24 * time.Hour,
		Methods: []string{http.MethodPost, http.MethodPatch},
	}
}

type cachedResponse struct {
	status  int
	headers http.Header
	body    []byte
	expiresAt time.Time
}

type store struct {
	mu    sync.Mutex
	items map[string]*cachedResponse
}

func (s *store) get(key string) (*cachedResponse, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.items[key]
	if !ok || time.Now().After(v.expiresAt) {
		delete(s.items, key)
		return nil, false
	}
	return v, true
}

func (s *store) set(key string, r *cachedResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = r
}

type recorder struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (r *recorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *recorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// New returns idempotency middleware using the provided options.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.Header == "" {
		opts.Header = DefaultOptions().Header
	}
	if opts.TTL == 0 {
		opts.TTL = DefaultOptions().TTL
	}
	if len(opts.Methods) == 0 {
		opts.Methods = DefaultOptions().Methods
	}

	allowed := make(map[string]struct{}, len(opts.Methods))
	for _, m := range opts.Methods {
		allowed[m] = struct{}{}
	}

	s := &store{items: make(map[string]*cachedResponse)}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := allowed[r.Method]; !ok {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get(opts.Header)
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			if cached, ok := s.get(key); ok {
				for k, vals := range cached.headers {
					for _, v := range vals {
						w.Header().Add(k, v)
					}
				}
				w.Header().Set("X-Idempotency-Replayed", "true")
				w.WriteHeader(cached.status)
				w.Write(cached.body) //nolint:errcheck
				return
			}

			rec := &recorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)

			s.set(key, &cachedResponse{
				status:    rec.status,
				headers:   w.Header().Clone(),
				body:      rec.body.Bytes(),
				expiresAt: time.Now().Add(opts.TTL),
			})
		})
	}
}
