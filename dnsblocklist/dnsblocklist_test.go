package dnsblocklist_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/crazywolf132/routeguard/dnsblocklist"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func fakeLookup(results []string, err error) func(string) ([]string, error) {
	return func(addr string) ([]string, error) {
		return results, err
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := dnsblocklist.DefaultOptions()
	if opts.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", opts.StatusCode)
	}
	if opts.Message == "" {
		t.Fatal("expected non-empty default message")
	}
}

func TestNew_NoDomains_PassesThrough(t *testing.T) {
	opts := dnsblocklist.DefaultOptions()
	opts.BlockedDomains = nil
	h := dnsblocklist.New(opts)(newTestHandler())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestNew_BlockedDomain_Returns403(t *testing.T) {
	opts := dnsblocklist.DefaultOptions()
	opts.BlockedDomains = []string{"evil.com"}
	opts.LookupAddr = fakeLookup([]string{"bot.evil.com."}, nil)
	h := dnsblocklist.New(opts)(newTestHandler())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestNew_AllowedDomain_PassesThrough(t *testing.T) {
	opts := dnsblocklist.DefaultOptions()
	opts.BlockedDomains = []string{"evil.com"}
	opts.LookupAddr = fakeLookup([]string{"good.example.com."}, nil)
	h := dnsblocklist.New(opts)(newTestHandler())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestNew_LookupError_PassesThrough(t *testing.T) {
	opts := dnsblocklist.DefaultOptions()
	opts.BlockedDomains = []string{"evil.com"}
	opts.LookupAddr = fakeLookup(nil, errors.New("no PTR record"))
	h := dnsblocklist.New(opts)(newTestHandler())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestNew_ExactDomainMatch_Blocked(t *testing.T) {
	opts := dnsblocklist.DefaultOptions()
	opts.BlockedDomains = []string{"spam.net"}
	opts.LookupAddr = fakeLookup([]string{"spam.net"}, nil)
	h := dnsblocklist.New(opts)(newTestHandler())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "5.6.7.8:9090"
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}
