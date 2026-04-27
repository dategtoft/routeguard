package logger

import (
	"log"
	"net/http"
	"os"
	"time"
)

// Logger holds the configuration for the logging middleware.
type Logger struct {
	log    *log.Logger
	prefix string
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// New creates a new Logger middleware with an optional prefix.
func New(prefix string) *Logger {
	if prefix == "" {
		prefix = "[routeguard]"
	}
	return &Logger{
		log:    log.New(os.Stdout, prefix+" ", log.LstdFlags),
		prefix: prefix,
	}
}

// Middleware returns an HTTP handler that logs each request.
func (l *Logger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		l.log.Printf("%s %s %d %s", r.Method, r.URL.Path, wrapped.statusCode, duration)
	})
}
