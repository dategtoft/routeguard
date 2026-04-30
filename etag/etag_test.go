package etag_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/routeguard/etag"
)

func newTestHandler(body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body)) //nolint:errcheck
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := etag.DefaultOptions()
	if opts.Weak {
		t.Error("expected Weak to be false by default")
	}
}

func TestNew_AddsETagHeader(t *testing.T) {
	mw := etag.New(etag.DefaultOptions())
	handler := mw(newTestHandler("hello world"))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("ETag") == "" {
		t.Error("expected ETag header to be set")
	}
}

func TestNew_Returns304OnMatch(t *testing.T) {
	mw := etag.New(etag.DefaultOptions())
	handler := mw(newTestHandler("hello world"))

	// First request to get the ETag.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)
	generatedETag := rec.Header().Get("ETag")

	// Second request with If-None-Match.
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("If-None-Match", generatedETag)
	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusNotModified {
		t.Errorf("expected 304, got %d", rec2.Code)
	}
}

func TestNew_WeakETag(t *testing.T) {
	opts := etag.Options{Weak: true}
	mw := etag.New(opts)
	handler := mw(newTestHandler("data"))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	e := rec.Header().Get("ETag")
	if len(e) < 2 || e[:2] != `W/` {
		t.Errorf("expected weak ETag, got %q", e)
	}
}

func TestNew_SkipsNonGET(t *testing.T) {
	mw := etag.New(etag.DefaultOptions())
	handler := mw(newTestHandler("body"))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("ETag") != "" {
		t.Error("expected no ETag header for POST requests")
	}
}

func TestNew_HeadRequest_NoBody(t *testing.T) {
	mw := etag.New(etag.DefaultOptions())
	handler := mw(newTestHandler("hello"))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Body.Len() != 0 {
		t.Error("expected empty body for HEAD request")
	}
	if rec.Header().Get("ETag") == "" {
		t.Error("expected ETag header to be set for HEAD request")
	}
}
