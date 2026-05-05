package maintenance_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/routeguard/maintenance"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := maintenance.DefaultOptions()
	if opts.Message == "" {
		t.Error("expected non-empty default message")
	}
	if opts.RetryAfter <= 0 {
		t.Error("expected positive default RetryAfter")
	}
	if !opts.JSONResponse {
		t.Error("expected JSONResponse to be true by default")
	}
}

func TestNew_Inactive_PassesThrough(t *testing.T) {
	m := maintenance.New(false, maintenance.DefaultOptions())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	m.Handler(newTestHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestNew_Active_Returns503(t *testing.T) {
	m := maintenance.New(true, maintenance.DefaultOptions())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	m.Handler(newTestHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestNew_Active_JSONBody(t *testing.T) {
	m := maintenance.New(true, maintenance.DefaultOptions())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	m.Handler(newTestHandler()).ServeHTTP(rec, req)

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("expected valid JSON body: %v", err)
	}
	if _, ok := body["error"]; !ok {
		t.Error("expected 'error' key in JSON response")
	}
}

func TestNew_Active_RetryAfterHeader(t *testing.T) {
	opts := maintenance.DefaultOptions()
	opts.RetryAfter = 60
	m := maintenance.New(true, opts)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	m.Handler(newTestHandler()).ServeHTTP(rec, req)
	if rec.Header().Get("Retry-After") != "60" {
		t.Errorf("expected Retry-After: 60, got %q", rec.Header().Get("Retry-After"))
	}
}

func TestNew_EnableDisable(t *testing.T) {
	m := maintenance.New(false, maintenance.DefaultOptions())

	m.Enable()
	if !m.Active() {
		t.Error("expected maintenance to be active after Enable()")
	}

	m.Disable()
	if m.Active() {
		t.Error("expected maintenance to be inactive after Disable()")
	}

	// Verify requests pass through after Disable
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	m.Handler(newTestHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 after Disable, got %d", rec.Code)
	}
}

func TestNew_PlainTextResponse(t *testing.T) {
	opts := maintenance.DefaultOptions()
	opts.JSONResponse = false
	opts.Message = "down for maintenance"
	m := maintenance.New(true, opts)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	m.Handler(newTestHandler()).ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
	if rec.Body.String() != "down for maintenance" {
		t.Errorf("unexpected body: %q", rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Errorf("unexpected Content-Type: %q", ct)
	}
}
