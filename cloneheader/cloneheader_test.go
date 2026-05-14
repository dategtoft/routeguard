package cloneheader_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaxron/routeguard/cloneheader"
)

func newCaptureHandler(captured *http.Header) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*captured = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := cloneheader.DefaultOptions()
	if len(opts.Rules) != 0 {
		t.Fatalf("expected no rules, got %d", len(opts.Rules))
	}
}

func TestNew_CopiesHeader(t *testing.T) {
	var captured http.Header
	h := cloneheader.New(cloneheader.Options{
		Rules: []cloneheader.Rule{
			{Source: "X-Forwarded-User", Destination: "X-User"},
		},
	}, newCaptureHandler(&captured))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-User", "alice")
	h.ServeHTTP(httptest.NewRecorder(), req)

	if got := captured.Get("X-User"); got != "alice" {
		t.Fatalf("expected X-User=alice, got %q", got)
	}
}

func TestNew_MissingSource_Skipped(t *testing.T) {
	var captured http.Header
	h := cloneheader.New(cloneheader.Options{
		Rules: []cloneheader.Rule{
			{Source: "X-Forwarded-User", Destination: "X-User"},
		},
	}, newCaptureHandler(&captured))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if got := captured.Get("X-User"); got != "" {
		t.Fatalf("expected X-User to be absent, got %q", got)
	}
}

func TestNew_NoOverwrite_KeepsExisting(t *testing.T) {
	var captured http.Header
	h := cloneheader.New(cloneheader.Options{
		Rules:     []cloneheader.Rule{{Source: "X-Src", Destination: "X-Dst"}},
		Overwrite: false,
	}, newCaptureHandler(&captured))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Src", "new")
	req.Header.Set("X-Dst", "original")
	h.ServeHTTP(httptest.NewRecorder(), req)

	if got := captured.Get("X-Dst"); got != "original" {
		t.Fatalf("expected X-Dst=original, got %q", got)
	}
}

func TestNew_Overwrite_ReplacesExisting(t *testing.T) {
	var captured http.Header
	h := cloneheader.New(cloneheader.Options{
		Rules:     []cloneheader.Rule{{Source: "X-Src", Destination: "X-Dst"}},
		Overwrite: true,
	}, newCaptureHandler(&captured))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Src", "new")
	req.Header.Set("X-Dst", "original")
	h.ServeHTTP(httptest.NewRecorder(), req)

	if got := captured.Get("X-Dst"); got != "new" {
		t.Fatalf("expected X-Dst=new, got %q", got)
	}
}

func TestNew_OriginalRequestUnmodified(t *testing.T) {
	h := cloneheader.New(cloneheader.Options{
		Rules: []cloneheader.Rule{{Source: "X-Src", Destination: "X-Dst"}},
	}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Src", "value")
	h.ServeHTTP(httptest.NewRecorder(), req)

	// Original request header must not have X-Dst injected.
	if got := req.Header.Get("X-Dst"); got != "" {
		t.Fatalf("original request was mutated: X-Dst=%q", got)
	}
}
