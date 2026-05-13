package querylog_test

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrickward/routeguard/querylog"
)

func BenchmarkQueryLog_NoParams(b *testing.B) {
	opts := querylog.DefaultOptions()
	opts.Logger = log.New(io.Discard, "", 0)
	mw := querylog.New(opts)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/path", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, req)
	}
}

func BenchmarkQueryLog_WithParams(b *testing.B) {
	opts := querylog.DefaultOptions()
	opts.Logger = log.New(io.Discard, "", 0)
	mw := querylog.New(opts)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/search?q=hello&page=1&sort=asc", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, req)
	}
}

func BenchmarkQueryLog_Redacted(b *testing.B) {
	opts := querylog.DefaultOptions()
	opts.Logger = log.New(io.Discard, "", 0)
	opts.RedactKeys = []string{"token", "api_key"}
	mw := querylog.New(opts)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/api?token=secret&user=alice", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, req)
	}
}
