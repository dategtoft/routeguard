// Package cache provides HTTP response caching middleware for Go HTTP routers.
package cache

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// Options holds configuration for the cache middleware.
type Options struct {
	// TTL is the duration for which a cached response is valid.
	TTL time.Duration
	// Methods defines which HTTP methods are eligible for caching.
	Methods []string
}

// DefaultOptions returns a sensible default Options configuration.
func DefaultOptions() Options {
	return Options{
		TTL:     60 * time.Second,
		Methods: []string{http.MethodGet, http.MethodHead},
	}
}

type entry struct {
	statuscode int
	headers    http.Header
	body       []byte
	expiresAt  time.Time
}

type store struct {
	mu      sync.RWMutex
	entries map[string]*entry
}

func (s *store) get(key string) (*entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[key]
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e, true
}

func (s *store) set(key string, e *entry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[key] = e
}

func allowedMethod(method string, methods []string) bool {
	for _, m := range methods {
		if m == method {
			return true
		}
	}
	return false
}

// New returns a caching middleware using the provided Options.
func New(opts Options) func(http.Handler) http.Handler {
	s := &store{entries: make(map[string]*entry)}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !allowedMethod(r.Method, opts.Methods) {
				next.ServeHTTP(w, r)
				return
			}
			key := r.URL.RequestURI()
			if e, ok := s.get(key); ok {
				for k, vals := range e.headers {
					for _, v := range vals {
						w.Header().Add(k, v)
					}
				}
				w.Header().Set("X-Cache", "HIT")
				w.WriteHeader(e.statuscode)
				w.Write(e.body) //nolint:errcheck
				return
			}
			rec := httptest.NewRecorder()
			next.ServeHTTP(rec, r)
			result := rec.Result()
			e := &entry{
				statuscode: result.StatusCode,
				headers:    result.Header.Clone(),
				body:       rec.Body.Bytes(),
				expiresAt:  time.Now().Add(opts.TTL),
			}
			s.set(key, e)
			for k, vals := range e.headers {
				for _, v := range vals {
					w.Header().Add(k, v)
				}
			}
			w.Header().Set("X-Cache", "MISS")
			w.WriteHeader(e.statuscode)
			w.Write(e.body) //nolint:errcheck
		})
	}
}
