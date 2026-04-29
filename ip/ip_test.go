package ip_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/routeguard/ip"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := ip.DefaultOptions()
	if opts.DeniedStatus != http.StatusForbidden {
		t.Errorf("expected 403, got %d", opts.DeniedStatus)
	}
	if opts.TrustProxy {
		t.Error("expected TrustProxy to be false")
	}
}

func TestAllowlist_AllowedIP(t *testing.T) {
	mw := ip.New(ip.Options{
		Allowlist:    []string{"192.168.1.1"},
		DeniedStatus: http.StatusForbidden,
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	rec := httptest.NewRecorder()
	mw(newTestHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestAllowlist_BlockedIP(t *testing.T) {
	mw := ip.New(ip.Options{
		Allowlist:    []string{"192.168.1.1"},
		DeniedStatus: http.StatusForbidden,
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rec := httptest.NewRecorder()
	mw(newTestHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestBlocklist_BlockedIP(t *testing.T) {
	mw := ip.New(ip.Options{
		Blocklist:    []string{"10.0.0.0/8"},
		DeniedStatus: http.StatusForbidden,
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.5.5.5:4321"
	rec := httptest.NewRecorder()
	mw(newTestHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestBlocklist_AllowedIP(t *testing.T) {
	mw := ip.New(ip.Options{
		Blocklist:    []string{"10.0.0.0/8"},
		DeniedStatus: http.StatusForbidden,
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:4321"
	rec := httptest.NewRecorder()
	mw(newTestHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestTrustProxy_XForwardedFor(t *testing.T) {
	mw := ip.New(ip.Options{
		Allowlist:    []string{"203.0.113.5"},
		DeniedStatus: http.StatusForbidden,
		TrustProxy:   true,
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:9999"
	req.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.1")
	rec := httptest.NewRecorder()
	mw(newTestHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 with trusted proxy header, got %d", rec.Code)
	}
}
