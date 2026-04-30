package rewrite_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/routeguard/rewrite"
)

func newCaptureHandler() (http.Handler, *string) {
	path := ""
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})
	return h, &path
}

func TestDefaultOptions(t *testing.T) {
	opts := rewrite.DefaultOptions()
	if len(opts.Rules) != 0 {
		t.Fatalf("expected 0 rules, got %d", len(opts.Rules))
	}
}

func TestNew_NoRules_PassThrough(t *testing.T) {
	h, path := newCaptureHandler()
	mw := rewrite.New(rewrite.DefaultOptions())(h)

	req := httptest.NewRequest(http.MethodGet, "/original/path", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if *path != "/original/path" {
		t.Fatalf("expected /original/path, got %s", *path)
	}
}

func TestNew_InternalRewrite(t *testing.T) {
	h, path := newCaptureHandler()

	opts := rewrite.DefaultOptions()
	if err := opts.AddRule(`^/old/(.*)$`, "/new/$1", false); err != nil {
		t.Fatal(err)
	}

	mw := rewrite.New(opts)(h)
	req := httptest.NewRequest(http.MethodGet, "/old/page", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if *path != "/new/page" {
		t.Fatalf("expected /new/page, got %s", *path)
	}
}

func TestNew_RedirectRule(t *testing.T) {
	h, _ := newCaptureHandler()

	opts := rewrite.DefaultOptions()
	if err := opts.AddRule(`^/legacy$`, "/current", true); err != nil {
		t.Fatal(err)
	}

	mw := rewrite.New(opts)(h)
	req := httptest.NewRequest(http.MethodGet, "/legacy", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("expected 301, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/current" {
		t.Fatalf("expected Location /current, got %s", loc)
	}
}

func TestNew_FirstMatchWins(t *testing.T) {
	h, path := newCaptureHandler()

	opts := rewrite.DefaultOptions()
	_ = opts.AddRule(`^/foo$`, "/first", false)
	_ = opts.AddRule(`^/foo$`, "/second", false)

	mw := rewrite.New(opts)(h)
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if *path != "/first" {
		t.Fatalf("expected /first, got %s", *path)
	}
}

func TestAddRule_InvalidPattern(t *testing.T) {
	opts := rewrite.DefaultOptions()
	if err := opts.AddRule(`[invalid`, "/x", false); err == nil {
		t.Fatal("expected error for invalid regex, got nil")
	}
}
