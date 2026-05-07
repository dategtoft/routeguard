package expiry_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jkellogg01/routeguard/expiry"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := expiry.DefaultOptions()
	if opts.MaxAge != 5*time.Minute {
		t.Errorf("expected MaxAge 5m, got %v", opts.MaxAge)
	}
}

func TestNew_SetsPublicMaxAge(t *testing.T) {
	mw := expiry.New(expiry.Options{MaxAge: 10 * time.Minute})
	rec := httptest.NewRecorder()
	mw(newTestHandler()).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	cc := rec.Header().Get("Cache-Control")
	if !strings.Contains(cc, "public") {
		t.Errorf("expected public in Cache-Control, got %q", cc)
	}
	if !strings.Contains(cc, "max-age=600") {
		t.Errorf("expected max-age=600 in Cache-Control, got %q", cc)
	}
	if rec.Header().Get("Expires") == "" {
		t.Error("expected Expires header to be set")
	}
}

func TestNew_PrivateScope(t *testing.T) {
	mw := expiry.New(expiry.Options{MaxAge: time.Minute, Private: true})
	rec := httptest.NewRecorder()
	mw(newTestHandler()).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	cc := rec.Header().Get("Cache-Control")
	if !strings.Contains(cc, "private") {
		t.Errorf("expected private in Cache-Control, got %q", cc)
	}
}

func TestNew_ImmutableDirective(t *testing.T) {
	mw := expiry.New(expiry.Options{MaxAge: time.Hour, Immutable: true})
	rec := httptest.NewRecorder()
	mw(newTestHandler()).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	cc := rec.Header().Get("Cache-Control")
	if !strings.Contains(cc, "immutable") {
		t.Errorf("expected immutable in Cache-Control, got %q", cc)
	}
}

func TestNew_NoStore(t *testing.T) {
	mw := expiry.New(expiry.Options{NoStore: true})
	rec := httptest.NewRecorder()
	mw(newTestHandler()).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	cc := rec.Header().Get("Cache-Control")
	if cc != "no-store" {
		t.Errorf("expected no-store, got %q", cc)
	}
	if rec.Header().Get("Expires") != "" {
		t.Error("expected no Expires header when NoStore is set")
	}
	if rec.Header().Get("Pragma") != "no-cache" {
		t.Error("expected Pragma: no-cache when NoStore is set")
	}
}
