package tracing_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/routeguard/tracing"
)

func newTestHandler(t *testing.T, checkCtx bool) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if checkCtx {
			if id := tracing.FromContext(r.Context()); id == "" {
				t.Error("expected trace ID in context, got empty string")
			}
		}
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := tracing.DefaultOptions()
	if opts.TraceHeader != "X-Trace-Id" {
		t.Errorf("expected X-Trace-Id, got %s", opts.TraceHeader)
	}
	if opts.SpanHeader != "X-Span-Id" {
		t.Errorf("expected X-Span-Id, got %s", opts.SpanHeader)
	}
}

func TestNew_GeneratesTraceAndSpanIDs(t *testing.T) {
	mw := tracing.New(tracing.DefaultOptions())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	mw(newTestHandler(t, true)).ServeHTTP(rec, req)

	if rec.Header().Get("X-Trace-Id") == "" {
		t.Error("expected X-Trace-Id header to be set")
	}
	if rec.Header().Get("X-Span-Id") == "" {
		t.Error("expected X-Span-Id header to be set")
	}
}

func TestNew_ReusesIncomingTraceID(t *testing.T) {
	mw := tracing.New(tracing.DefaultOptions())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Trace-Id", "existing-trace-id")

	mw(newTestHandler(t, false)).ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Trace-Id"); got != "existing-trace-id" {
		t.Errorf("expected existing-trace-id, got %s", got)
	}
}

func TestNew_SpanIDAlwaysNew(t *testing.T) {
	mw := tracing.New(tracing.DefaultOptions())

	rec1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(newTestHandler(t, false)).ServeHTTP(rec1, req1)

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(newTestHandler(t, false)).ServeHTTP(rec2, req2)

	if rec1.Header().Get("X-Span-Id") == rec2.Header().Get("X-Span-Id") {
		t.Error("expected different span IDs across requests")
	}
}

func TestNew_CustomHeader(t *testing.T) {
	opts := tracing.DefaultOptions()
	opts.TraceHeader = "X-Request-Trace"
	opts.SpanHeader = "X-Request-Span"

	mw := tracing.New(opts)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	mw(newTestHandler(t, false)).ServeHTTP(rec, req)

	if rec.Header().Get("X-Request-Trace") == "" {
		t.Error("expected X-Request-Trace header to be set")
	}
	if rec.Header().Get("X-Request-Span") == "" {
		t.Error("expected X-Request-Span header to be set")
	}
}

func TestNew_CustomGenerator(t *testing.T) {
	opts := tracing.DefaultOptions()
	opts.Generator = func() string { return "fixed-id" }

	mw := tracing.New(opts)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	mw(newTestHandler(t, false)).ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Trace-Id"); got != "fixed-id" {
		t.Errorf("expected fixed-id, got %s", got)
	}
}
