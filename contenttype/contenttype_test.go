package contenttype_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/username/routeguard/contenttype"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := contenttype.DefaultOptions()
	if len(opts.SkipMethods) == 0 {
		t.Fatal("expected default skip methods to be non-empty")
	}
}

func TestNew_AllowedContentType_Passes(t *testing.T) {
	mw := contenttype.New(contenttype.Options{
		Allowed:     []string{"application/json"},
		SkipMethods: []string{http.MethodGet},
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mw(newTestHandler()).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestNew_DisallowedContentType_Returns415(t *testing.T) {
	mw := contenttype.New(contenttype.Options{
		Allowed:     []string{"application/json"},
		SkipMethods: []string{http.MethodGet},
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`data`))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()

	mw(newTestHandler()).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415, got %d", rec.Code)
	}
}

func TestNew_SkippedMethod_NotChecked(t *testing.T) {
	mw := contenttype.New(contenttype.Options{
		Allowed:     []string{"application/json"},
		SkipMethods: []string{http.MethodGet},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// No Content-Type header set — should still pass.
	rec := httptest.NewRecorder()

	mw(newTestHandler()).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestNew_ResponseTypeSet(t *testing.T) {
	mw := contenttype.New(contenttype.Options{
		ResponseType: "application/json; charset=utf-8",
		SkipMethods:  []string{http.MethodGet},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	mw(newTestHandler()).ServeHTTP(rec, req)

	if got := rec.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("unexpected Content-Type: %q", got)
	}
}

func TestNew_ContentTypeWithCharset_Passes(t *testing.T) {
	mw := contenttype.New(contenttype.Options{
		Allowed:     []string{"application/json"},
		SkipMethods: []string{http.MethodGet},
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	rec := httptest.NewRecorder()

	mw(newTestHandler()).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
