package referrer_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/routeguard/referrer"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := referrer.DefaultOptions()
	if opts.Policy != "strict-origin-when-cross-origin" {
		t.Errorf("expected default policy, got %q", opts.Policy)
	}
	if opts.BlockEmpty {
		t.Error("expected BlockEmpty to be false by default")
	}
}

func TestNew_SetsPolicyHeader(t *testing.T) {
	mw := referrer.New(referrer.DefaultOptions())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	mw(newTestHandler()).ServeHTTP(rec, req)

	if got := rec.Header().Get("Referrer-Policy"); got != "strict-origin-when-cross-origin" {
		t.Errorf("expected Referrer-Policy header, got %q", got)
	}
}

func TestNew_CustomPolicy(t *testing.T) {
	mw := referrer.New(referrer.Options{Policy: "no-referrer"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	mw(newTestHandler()).ServeHTTP(rec, req)

	if got := rec.Header().Get("Referrer-Policy"); got != "no-referrer" {
		t.Errorf("expected no-referrer policy, got %q", got)
	}
}

func TestNew_AllowedHost_Passes(t *testing.T) {
	mw := referrer.New(referrer.Options{
		AllowedHosts: []string{"trusted.example.com"},
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Referer", "https://trusted.example.com/page")

	mw(newTestHandler()).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestNew_DisallowedHost_Returns403(t *testing.T) {
	mw := referrer.New(referrer.Options{
		AllowedHosts: []string{"trusted.example.com"},
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Referer", "https://evil.example.com/page")

	mw(newTestHandler()).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestNew_EmptyReferer_AllowedByDefault(t *testing.T) {
	mw := referrer.New(referrer.Options{
		AllowedHosts: []string{"trusted.example.com"},
		BlockEmpty:   false,
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	mw(newTestHandler()).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for empty referer, got %d", rec.Code)
	}
}

func TestNew_EmptyReferer_BlockedWhenConfigured(t *testing.T) {
	mw := referrer.New(referrer.Options{
		AllowedHosts: []string{"trusted.example.com"},
		BlockEmpty:   true,
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	mw(newTestHandler()).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for empty referer with BlockEmpty, got %d", rec.Code)
	}
}
