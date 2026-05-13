package allowlist_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrickward/routeguard/allowlist"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := allowlist.DefaultOptions()
	if len(opts.Methods) == 0 {
		t.Fatal("expected default methods to be non-empty")
	}
	if opts.Message == "" {
		t.Fatal("expected default message to be non-empty")
	}
}

func TestNew_AllowedMethod_Passes(t *testing.T) {
	mw := allowlist.New(allowlist.Options{
		Methods: []string{http.MethodGet, http.MethodPost},
	})
	handler := mw(newTestHandler())

	for _, method := range []string{http.MethodGet, http.MethodPost} {
		req := httptest.NewRequest(method, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("%s: expected 200, got %d", method, rec.Code)
		}
	}
}

func TestNew_BlockedMethod_Returns405(t *testing.T) {
	mw := allowlist.New(allowlist.Options{
		Methods: []string{http.MethodGet},
	})
	handler := mw(newTestHandler())

	blocked := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}
	for _, method := range blocked {
		req := httptest.NewRequest(method, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: expected 405, got %d", method, rec.Code)
		}
	}
}

func TestNew_CustomMessage(t *testing.T) {
	mw := allowlist.New(allowlist.Options{
		Methods:  []string{http.MethodGet},
		Message:  "nope",
	})
	handler := mw(newTestHandler())

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "nope\n" {
		t.Errorf("unexpected body: %q", body)
	}
}

func TestNew_EmptyOptions_UsesDefaults(t *testing.T) {
	mw := allowlist.New(allowlist.Options{})
	handler := mw(newTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
