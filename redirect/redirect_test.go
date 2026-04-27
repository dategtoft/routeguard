package redirect_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/routeguard/redirect"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := redirect.DefaultOptions()
	if opts.StatusCode != http.StatusMovedPermanently {
		t.Errorf("expected status 301, got %d", opts.StatusCode)
	}
	if opts.HTTPSOnly {
		t.Error("expected HTTPSOnly to be false by default")
	}
	if opts.TrailingSlash != "" {
		t.Errorf("expected empty TrailingSlash, got %q", opts.TrailingSlash)
	}
}

func TestHTTPSOnly_RedirectsHTTP(t *testing.T) {
	opts := redirect.DefaultOptions()
	opts.HTTPSOnly = true
	mw := redirect.New(opts)(newTestHandler())

	req := httptest.NewRequest(http.MethodGet, "http://example.com/page", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusMovedPermanently {
		t.Errorf("expected 301, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc != "https://example.com/page" {
		t.Errorf("unexpected Location: %s", loc)
	}
}

func TestHTTPSOnly_PassesThroughHTTPS(t *testing.T) {
	opts := redirect.DefaultOptions()
	opts.HTTPSOnly = true
	mw := redirect.New(opts)(newTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/page", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestTrailingSlash_Add(t *testing.T) {
	opts := redirect.DefaultOptions()
	opts.TrailingSlash = "add"
	mw := redirect.New(opts)(newTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/about", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusMovedPermanently {
		t.Errorf("expected 301, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/about/" {
		t.Errorf("expected /about/, got %s", loc)
	}
}

func TestTrailingSlash_Remove(t *testing.T) {
	opts := redirect.DefaultOptions()
	opts.TrailingSlash = "remove"
	mw := redirect.New(opts)(newTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/about/", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusMovedPermanently {
		t.Errorf("expected 301, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/about" {
		t.Errorf("expected /about, got %s", loc)
	}
}

func TestNoRedirect_PassesThrough(t *testing.T) {
	mw := redirect.New(redirect.DefaultOptions())(newTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/about", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
