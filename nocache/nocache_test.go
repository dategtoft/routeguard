package nocache_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrickward/routeguard/nocache"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := nocache.DefaultOptions()
	if !opts.Pragma {
		t.Error("expected Pragma to be true by default")
	}
	if !opts.Expires {
		t.Error("expected Expires to be true by default")
	}
}

func TestNew_SetsNoCacheHeaders(t *testing.T) {
	h := nocache.New(nocache.DefaultOptions())(newTestHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	cc := rec.Header().Get("Cache-Control")
	if cc == "" {
		t.Fatal("expected Cache-Control header to be set")
	}
	for _, directive := range []string{"no-store", "no-cache", "must-revalidate"} {
		if !contains(cc, directive) {
			t.Errorf("expected Cache-Control to contain %q, got %q", directive, cc)
		}
	}
}

func TestNew_SetsPragmaHeader(t *testing.T) {
	h := nocache.New(nocache.DefaultOptions())(newTestHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Header().Get("Pragma"); got != "no-cache" {
		t.Errorf("expected Pragma: no-cache, got %q", got)
	}
}

func TestNew_SetsExpiresHeader(t *testing.T) {
	h := nocache.New(nocache.DefaultOptions())(newTestHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Header().Get("Expires"); got != "0" {
		t.Errorf("expected Expires: 0, got %q", got)
	}
}

func TestNew_PragmaDisabled(t *testing.T) {
	opts := nocache.DefaultOptions()
	opts.Pragma = false
	h := nocache.New(opts)(newTestHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Header().Get("Pragma"); got != "" {
		t.Errorf("expected Pragma header to be absent, got %q", got)
	}
}

func TestNew_ExpiresDisabled(t *testing.T) {
	opts := nocache.DefaultOptions()
	opts.Expires = false
	h := nocache.New(opts)(newTestHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Header().Get("Expires"); got != "" {
		t.Errorf("expected Expires header to be absent, got %q", got)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsRune(s, sub))
}

func containsRune(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
