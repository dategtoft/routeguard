package stripprefix_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/routeguard/stripprefix"
)

func newCaptureHandler() (http.Handler, *string) {
	var captured string
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})
	return h, &captured
}

func TestDefaultOptions(t *testing.T) {
	opts := stripprefix.DefaultOptions()
	if opts.Prefix != "" {
		t.Errorf("expected empty prefix, got %q", opts.Prefix)
	}
	if opts.RedirectOnMismatch {
		t.Error("expected RedirectOnMismatch to be false")
	}
}

func TestNew_NoPrefix_PassesThrough(t *testing.T) {
	h, captured := newCaptureHandler()
	mw := stripprefix.New(stripprefix.DefaultOptions())(h)

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if *captured != "/api/users" {
		t.Errorf("expected /api/users, got %q", *captured)
	}
}

func TestNew_StripsPrefix(t *testing.T) {
	h, captured := newCaptureHandler()
	mw := stripprefix.New(stripprefix.Options{Prefix: "/api"})(h)

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if *captured != "/users" {
		t.Errorf("expected /users, got %q", *captured)
	}
}

func TestNew_MismatchPassesThrough(t *testing.T) {
	h, captured := newCaptureHandler()
	mw := stripprefix.New(stripprefix.Options{Prefix: "/api", RedirectOnMismatch: false})(h)

	req := httptest.NewRequest(http.MethodGet, "/other/path", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if *captured != "/other/path" {
		t.Errorf("expected /other/path, got %q", *captured)
	}
}

func TestNew_MismatchReturns404(t *testing.T) {
	h, _ := newCaptureHandler()
	mw := stripprefix.New(stripprefix.Options{Prefix: "/api", RedirectOnMismatch: true})(h)

	req := httptest.NewRequest(http.MethodGet, "/other/path", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestNew_PrefixWithTrailingSlash(t *testing.T) {
	h, captured := newCaptureHandler()
	mw := stripprefix.New(stripprefix.Options{Prefix: "/api/"})(h)

	req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if *captured != "/orders" {
		t.Errorf("expected /orders, got %q", *captured)
	}
}

func TestNew_ExactPrefixMatch(t *testing.T) {
	h, captured := newCaptureHandler()
	mw := stripprefix.New(stripprefix.Options{Prefix: "/api"})(h)

	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if *captured != "/" {
		t.Errorf("expected /, got %q", *captured)
	}
}
