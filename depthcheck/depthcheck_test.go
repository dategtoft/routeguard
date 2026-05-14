package depthcheck_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jwtly10/routeguard/depthcheck"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := depthcheck.DefaultOptions()
	if opts.MaxDepth != 10 {
		t.Errorf("expected MaxDepth 10, got %d", opts.MaxDepth)
	}
	if opts.Message == "" {
		t.Error("expected non-empty default Message")
	}
}

func TestNew_NonJSON_PassesThrough(t *testing.T) {
	mw := depthcheck.New(depthcheck.DefaultOptions())
	h := mw(newTestHandler())

	body := strings.NewReader(`{"a":{"b":"c"}}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	// no Content-Type header
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestNew_ShallowJSON_Passes(t *testing.T) {
	mw := depthcheck.New(depthcheck.Options{MaxDepth: 3})
	h := mw(newTestHandler())

	body := strings.NewReader(`{"a":{"b":"c"}}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestNew_DeeplyNestedJSON_Returns400(t *testing.T) {
	mw := depthcheck.New(depthcheck.Options{MaxDepth: 2})
	h := mw(newTestHandler())

	body := strings.NewReader(`{"a":{"b":{"c":"d"}}}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestNew_NilBody_PassesThrough(t *testing.T) {
	mw := depthcheck.New(depthcheck.DefaultOptions())
	h := mw(newTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestNew_CustomMessage(t *testing.T) {
	mw := depthcheck.New(depthcheck.Options{MaxDepth: 1, Message: "too deep"})
	h := mw(newTestHandler())

	body := strings.NewReader(`{"a":{"b":"c"}}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if !strings.Contains(rec.Body.String(), "too deep") {
		t.Errorf("expected custom message in body, got %q", rec.Body.String())
	}
}
