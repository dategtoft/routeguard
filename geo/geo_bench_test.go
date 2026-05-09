package geo_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/routeguard/geo"
)

func BenchmarkGeo_Allowlist_Hit(b *testing.B) {
	opts := geo.DefaultOptions()
	opts.Lookup = staticLookup("US")
	opts.Allowlist = []string{"US", "CA", "GB"}
	mw := geo.New(opts)
	handler := mw(newTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkGeo_Allowlist_Miss(b *testing.B) {
	opts := geo.DefaultOptions()
	opts.Lookup = staticLookup("CN")
	opts.Allowlist = []string{"US", "CA", "GB"}
	mw := geo.New(opts)
	handler := mw(newTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkGeo_NilLookup(b *testing.B) {
	opts := geo.DefaultOptions()
	mw := geo.New(opts)
	handler := mw(newTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}
