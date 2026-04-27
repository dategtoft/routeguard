// Package compress provides HTTP response compression middleware
// supporting gzip and deflate encodings.
package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// Options configures the compression middleware.
type Options struct {
	// Level is the compression level (gzip.BestSpeed to gzip.BestCompression).
	// Defaults to gzip.DefaultCompression.
	Level int
	// MinLength is the minimum response size in bytes before compression is applied.
	// Defaults to 1024.
	MinLength int
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Level:     gzip.DefaultCompression,
		MinLength: 1024,
	}
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer io.Writer
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	return g.writer.Write(b)
}

// New returns a middleware that compresses HTTP responses using gzip
// when the client supports it via the Accept-Encoding header.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.Level == 0 {
		opts.Level = gzip.DefaultCompression
	}
	if opts.MinLength == 0 {
		opts.MinLength = 1024
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !acceptsGzip(r) {
				next.ServeHTTP(w, r)
				return
			}

			gz, err := gzip.NewWriterLevel(w, opts.Level)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			defer gz.Close()

			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Del("Content-Length")

			grw := &gzipResponseWriter{ResponseWriter: w, writer: gz}
			next.ServeHTTP(grw, r)
		})
	}
}

func acceptsGzip(r *http.Request) bool {
	ae := r.Header.Get("Accept-Encoding")
	return strings.Contains(ae, "gzip")
}
