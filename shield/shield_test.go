package shield_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joeydtaylor/routeguard/shield"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := shield.DefaultOptions()
	if opts.ContentSecurityPolicy == "" {
		t.Error("expected non-empty ContentSecurityPolicy")
	}
	if opts.XFrameOptions == "" {
		t.Error("expected non-empty XFrameOptions")
	}
}

func TestNew_SetsDefaultSecurityHeaders(t *testing.T) {
	h := shield.New(shield.DefaultOptions())(newTestHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	checks := map[string]string{
		"Content-Security-Policy": "default-src 'self'",
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "DENY",
		"X-XSS-Protection":        "1; mode=block",
		"Referrer-Policy":         "strict-origin-when-cross-origin",
	}
	for header, want := range checks {
		if got := rec.Header().Get(header); got != want {
			t.Errorf("%s: got %q, want %q", header, got, want)
		}
	}
}

func TestNew_CustomCSP(t *testing.T) {
	opts := shield.DefaultOptions()
	opts.ContentSecurityPolicy = "default-src 'none'; script-src 'self'"
	h := shield.New(opts)(newTestHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Header().Get("Content-Security-Policy"); got != opts.ContentSecurityPolicy {
		t.Errorf("got %q, want %q", got, opts.ContentSecurityPolicy)
	}
}

func TestNew_PermissionsPolicyOmittedByDefault(t *testing.T) {
	h := shield.New(shield.DefaultOptions())(newTestHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Header().Get("Permissions-Policy"); got != "" {
		t.Errorf("expected Permissions-Policy to be absent, got %q", got)
	}
}

func TestNew_PermissionsPolicySet(t *testing.T) {
	opts := shield.DefaultOptions()
	opts.PermissionsPolicy = "geolocation=(), microphone=()"
	h := shield.New(opts)(newTestHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Header().Get("Permissions-Policy"); got != opts.PermissionsPolicy {
		t.Errorf("got %q, want %q", got, opts.PermissionsPolicy)
	}
}

func TestNew_EmptyOptionsOmitsHeaders(t *testing.T) {
	h := shield.New(shield.Options{})(newTestHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	for _, header := range []string{
		"Content-Security-Policy",
		"X-Content-Type-Options",
		"X-Frame-Options",
		"X-XSS-Protection",
		"Referrer-Policy",
		"Permissions-Policy",
	} {
		if got := rec.Header().Get(header); got != "" {
			t.Errorf("expected %s to be absent, got %q", header, got)
		}
	}
}
