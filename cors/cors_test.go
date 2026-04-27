package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestNew_DefaultOptions_WildcardOrigin(t *testing.T) {
	handler := New(DefaultOptions())(newTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected '*', got %q", got)
	}
}

func TestNew_SpecificOrigin_Allowed(t *testing.T) {
	opts := Options{
		AllowedOrigins: []string{"http://trusted.com"},
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{"Content-Type"},
	}
	handler := New(opts)(newTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://trusted.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "http://trusted.com" {
		t.Errorf("expected 'http://trusted.com', got %q", got)
	}
}

func TestNew_SpecificOrigin_NotAllowed(t *testing.T) {
	opts := Options{
		AllowedOrigins: []string{"http://trusted.com"},
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{"Content-Type"},
	}
	handler := New(opts)(newTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://evil.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected empty origin header, got %q", got)
	}
}

func TestNew_PreflightRequest(t *testing.T) {
	handler := New(DefaultOptions())(newTestHandler())

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 for preflight, got %d", rr.Code)
	}
}

func TestNew_AllowCredentials(t *testing.T) {
	opts := Options{
		AllowedOrigins:   []string{"http://trusted.com"},
		AllowedMethods:   []string{"GET"},
		AllowedHeaders:   []string{"Authorization"},
		AllowCredentials: true,
	}
	handler := New(opts)(newTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://trusted.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("expected 'true', got %q", got)
	}
}
