package ratelimit

import (
	"net/http"
	"sync"
	"time"
)

// Limiter holds the rate limiting configuration and state per client IP.
type Limiter struct {
	mu       sync.Mutex
	clients  map[string]*bucket
	rate     int
	window   time.Duration
}

type bucket struct {
	count    int
	resetAt  time.Time
}

// New creates a new Limiter that allows up to `rate` requests per `window` duration.
func New(rate int, window time.Duration) *Limiter {
	return &Limiter{
		clients: make(map[string]*bucket),
		rate:    rate,
		window:  window,
	}
}

// Allow checks whether the given key (e.g. client IP) is within the rate limit.
// Returns true if the request is allowed, false if it exceeds the limit.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, exists := l.clients[key]
	if !exists || now.After(b.resetAt) {
		l.clients[key] = &bucket{
			count:   1,
			resetAt: now.Add(l.window),
		}
		return true
	}

	if b.count >= l.rate {
		return false
	}

	b.count++
	return true
}

// Middleware returns an http.Handler middleware that enforces rate limiting by client IP.
func (l *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !l.Allow(ip) {
			http.Error(w, "429 Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// clientIP extracts the client IP from the request, respecting X-Forwarded-For.
func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return fwd
	}
	return r.RemoteAddr
}
