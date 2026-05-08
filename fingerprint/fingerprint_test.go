package fingerprint_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joeydtaylor/routeguard/fingerprint"
)

func newTestHandler(t *testing.T, checkCtx bool) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if checkCtx {
			fp := fingerprint.FromContext(r.Context())
			if fp == "" {
				t.Error("expected fingerprint in context, got empty string")
			}
		}
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := fingerprint.DefaultOptions()
	if opts.Header != "X-Request-Fingerprint" {
		t.Errorf("expected default header X-Request-Fingerprint, got %s", opts.Header)
	}
	if !opts.IncludeIP {
		t.Error("expected IncludeIP to be true by default")
	}
	if !opts.IncludeUserAgent {
		t.Error("expected IncludeUserAgent to be true by default")
	}
}

func TestNew_FingerprintHeaderPresent(t *testing.T) {
	mw := fingerprint.New(fingerprint.DefaultOptions())
	handler := mw(newTestHandler(t, false))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:5000"
	req.Header.Set("User-Agent", "test-agent")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	fp := rec.Header().Get("X-Request-Fingerprint")
	if fp == "" {
		t.Fatal("expected X-Request-Fingerprint header to be set")
	}
	if len(fp) != 64 {
		t.Errorf("expected 64-char hex SHA-256, got length %d", len(fp))
	}
}

func TestNew_FingerprintStoredInContext(t *testing.T) {
	mw := fingerprint.New(fingerprint.DefaultOptions())
	handler := mw(newTestHandler(t, true))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:9000"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
}

func TestNew_SameInputSameFingerprint(t *testing.T) {
	mw := fingerprint.New(fingerprint.DefaultOptions())

	makeReq := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.2:1234"
		req.Header.Set("User-Agent", "go-test")
		rec := httptest.NewRecorder()
		mw(newTestHandler(t, false)).ServeHTTP(rec, req)
		return rec
	}

	a := makeReq().Header.Get("X-Request-Fingerprint")
	b := makeReq().Header.Get("X-Request-Fingerprint")
	if a != b {
		t.Errorf("expected identical fingerprints, got %s vs %s", a, b)
	}
}

func TestNew_DifferentIPsDifferentFingerprints(t *testing.T) {
	mw := fingerprint.New(fingerprint.DefaultOptions())

	handler := mw(newTestHandler(t, false))

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "1.2.3.4:80"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "5.6.7.8:80"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec1.Header().Get("X-Request-Fingerprint") == rec2.Header().Get("X-Request-Fingerprint") {
		t.Error("expected different fingerprints for different IPs")
	}
}

func TestNew_ExtraHeaders(t *testing.T) {
	opts := fingerprint.DefaultOptions()
	opts.ExtraHeaders = []string{"X-Custom-ID"}
	mw := fingerprint.New(opts)
	handler := mw(newTestHandler(t, false))

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "1.1.1.1:80"
	req1.Header.Set("X-Custom-ID", "aaa")
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "1.1.1.1:80"
	req2.Header.Set("X-Custom-ID", "bbb")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec1.Header().Get("X-Request-Fingerprint") == rec2.Header().Get("X-Request-Fingerprint") {
		t.Error("expected different fingerprints for different X-Custom-ID values")
	}
}

func TestNew_CustomHeader(t *testing.T) {
	opts := fingerprint.DefaultOptions()
	opts.Header = "X-FP"
	mw := fingerprint.New(opts)
	handler := mw(newTestHandler(t, false))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-FP") == "" {
		t.Error("expected X-FP header to be set")
	}
}
