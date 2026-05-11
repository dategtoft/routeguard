package cachecontrol_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joeydotdev/routeguard/cachecontrol"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := cachecontrol.DefaultOptions()
	if opts.MaxAge != 60 {
		t.Fatalf("expected MaxAge=60, got %d", opts.MaxAge)
	}
	if opts.Private {
		t.Fatal("expected Private=false")
	}
}

func TestNew_PublicMaxAge(t *testing.T) {
	h := cachecontrol.New(cachecontrol.Options{MaxAge: 3600})(newTestHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	got := rec.Header().Get("Cache-Control")
	if got != "public, max-age=3600" {
		t.Fatalf("unexpected Cache-Control: %q", got)
	}
}

func TestNew_Private(t *testing.T) {
	h := cachecontrol.New(cachecontrol.Options{Private: true})(newTestHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	got := rec.Header().Get("Cache-Control")
	if got != "private" {
		t.Fatalf("unexpected Cache-Control: %q", got)
	}
}

func TestNew_NoStore(t *testing.T) {
	h := cachecontrol.New(cachecontrol.Options{NoStore: true})(newTestHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	got := rec.Header().Get("Cache-Control")
	if got != "no-store" {
		t.Fatalf("unexpected Cache-Control: %q", got)
	}
}

func TestNew_Immutable(t *testing.T) {
	h := cachecontrol.New(cachecontrol.Options{MaxAge: 31536000, Immutable: true})(newTestHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	got := rec.Header().Get("Cache-Control")
	expected := "public, max-age=31536000, immutable"
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestNew_SkipMethod(t *testing.T) {
	h := cachecontrol.New(cachecontrol.Options{
		MaxAge:      60,
		SkipMethods: []string{"POST"},
	})(newTestHandler())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))

	if got := rec.Header().Get("Cache-Control"); got != "" {
		t.Fatalf("expected no Cache-Control header for POST, got %q", got)
	}

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec2.Header().Get("Cache-Control"); got == "" {
		t.Fatal("expected Cache-Control header for GET")
	}
}

func TestNew_NoCacheWithMustRevalidate(t *testing.T) {
	h := cachecontrol.New(cachecontrol.Options{
		NoCache:        true,
		MustRevalidate: true,
	})(newTestHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	got := rec.Header().Get("Cache-Control")
	expected := "no-cache, public, must-revalidate"
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}
