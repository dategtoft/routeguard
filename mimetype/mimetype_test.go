package mimetype_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrickward/routeguard/mimetype"
)

func newTestHandler(contentType string, status int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		w.WriteHeader(status)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := mimetype.DefaultOptions()
	if len(opts.Allowed) == 0 {
		t.Fatal("expected non-empty default allowed list")
	}
}

func TestNew_AllowedMIME_Passes(t *testing.T) {
	mw := mimetype.New(mimetype.Options{
		Allowed: []string{"application/json"},
	})
	handler := mw(newTestHandler("application/json", http.StatusOK))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusNotAcceptable {
		t.Errorf("expected request to pass, got 406")
	}
}

func TestNew_DisallowedMIME_Returns406(t *testing.T) {
	mw := mimetype.New(mimetype.Options{
		Allowed: []string{"application/json"},
	})
	handler := mw(newTestHandler("text/xml", http.StatusOK))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotAcceptable {
		t.Errorf("expected 406, got %d", rec.Code)
	}
}

func TestNew_WildcardMIME_Passes(t *testing.T) {
	mw := mimetype.New(mimetype.Options{
		Allowed: []string{"image/*"},
	})
	handler := mw(newTestHandler("image/png", http.StatusOK))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusNotAcceptable {
		t.Errorf("expected image/png to match image/*, got 406")
	}
}

func TestNew_MIMEWithParams_Passes(t *testing.T) {
	mw := mimetype.New(mimetype.Options{
		Allowed: []string{"text/html"},
	})
	handler := mw(newTestHandler("text/html; charset=utf-8", http.StatusOK))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusNotAcceptable {
		t.Errorf("expected text/html with params to pass, got 406")
	}
}

func TestNew_EmptyContentType_Passes(t *testing.T) {
	mw := mimetype.New(mimetype.Options{
		Allowed: []string{"application/json"},
	})
	handler := mw(newTestHandler("", http.StatusOK))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusNotAcceptable {
		t.Errorf("expected empty Content-Type to pass through, got 406")
	}
}

func TestNew_CustomOnReject(t *testing.T) {
	called := false
	mw := mimetype.New(mimetype.Options{
		Allowed: []string{"application/json"},
		OnReject: func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusUnsupportedMediaType)
		},
	})
	handler := mw(newTestHandler("text/xml", http.StatusOK))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("expected custom OnReject to be called")
	}
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("expected 415, got %d", rec.Code)
	}
}
