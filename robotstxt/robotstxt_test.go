package robotstxt_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joeydtaylor/routeguard/robotstxt"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := robotstxt.DefaultOptions()
	if opts.Path != "/robots.txt" {
		t.Fatalf("expected /robots.txt, got %s", opts.Path)
	}
	if opts.Content == "" {
		t.Fatal("expected non-empty default content")
	}
	if opts.BlockBots {
		t.Fatal("expected BlockBots to be false by default")
	}
}

func TestNew_ServesRobotsTxt(t *testing.T) {
	mw := robotstxt.New(robotstxt.DefaultOptions())
	h := mw(newTestHandler())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Fatalf("unexpected Content-Type: %s", ct)
	}
	if body := rr.Body.String(); body == "" {
		t.Fatal("expected non-empty robots.txt body")
	}
}

func TestNew_PassesThroughNonRobotsPaths(t *testing.T) {
	mw := robotstxt.New(robotstxt.DefaultOptions())
	h := mw(newTestHandler())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if rr.Body.String() != "ok" {
		t.Fatal("expected downstream handler to respond")
	}
}

func TestNew_BlockBot_Returns403(t *testing.T) {
	opts := robotstxt.DefaultOptions()
	opts.BlockBots = true
	mw := robotstxt.New(opts)
	h := mw(newTestHandler())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/page", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1)")
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestNew_LegitUser_NotBlocked(t *testing.T) {
	opts := robotstxt.DefaultOptions()
	opts.BlockBots = true
	mw := robotstxt.New(opts)
	h := mw(newTestHandler())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/page", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120")
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestNew_CustomContent(t *testing.T) {
	opts := robotstxt.DefaultOptions()
	opts.Content = "User-agent: *\nDisallow: /private/"
	mw := robotstxt.New(opts)
	h := mw(newTestHandler())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	h.ServeHTTP(rr, req)

	if body := rr.Body.String(); body != "User-agent: *\nDisallow: /private/" {
		t.Fatalf("unexpected body: %s", body)
	}
}
