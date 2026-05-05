package vary_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrickward/routeguard/vary"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := vary.DefaultOptions()
	if len(opts.Headers) == 0 {
		t.Fatal("expected at least one default header")
	}
}

func TestNew_SetsVaryHeader(t *testing.T) {
	mw := vary.New(vary.Options{Headers: []string{"Accept-Encoding"}})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	mw(newTestHandler()).ServeHTTP(rec, req)

	if got := rec.Header().Get("Vary"); got != "Accept-Encoding" {
		t.Errorf("expected Vary: Accept-Encoding, got %q", got)
	}
}

func TestNew_MultipleHeaders(t *testing.T) {
	mw := vary.New(vary.Options{Headers: []string{"Accept-Encoding", "Accept-Language"}})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	mw(newTestHandler()).ServeHTTP(rec, req)

	got := rec.Header().Get("Vary")
	if got != "Accept-Encoding, Accept-Language" {
		t.Errorf("unexpected Vary header: %q", got)
	}
}

func TestNew_DeduplicatesHeaders(t *testing.T) {
	mw := vary.New(vary.Options{Headers: []string{"Accept-Encoding", "accept-encoding"}})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	mw(newTestHandler()).ServeHTTP(rec, req)

	got := rec.Header().Get("Vary")
	if got != "Accept-Encoding" {
		t.Errorf("expected deduplicated header, got %q", got)
	}
}

func TestNew_MergesWithExistingVary(t *testing.T) {
	mw := vary.New(vary.Options{Headers: []string{"Accept-Language"}})
	rec := httptest.NewRecorder()
	rec.Header().Set("Vary", "Accept-Encoding")
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Wrap a handler that pre-sets a Vary header.
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "Accept-Encoding")
		w.WriteHeader(http.StatusOK)
	})

	mw(inner).ServeHTTP(rec, req)

	got := rec.Header().Get("Vary")
	if got != "Accept-Encoding, Accept-Language" {
		t.Errorf("expected merged Vary header, got %q", got)
	}
}

func TestNew_WildcardShortCircuit(t *testing.T) {
	mw := vary.New(vary.Options{Headers: []string{"*"}})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	mw(newTestHandler()).ServeHTTP(rec, req)

	if got := rec.Header().Get("Vary"); got != "*" {
		t.Errorf("expected Vary: *, got %q", got)
	}
}

func TestNew_ExistingWildcard_NotOverwritten(t *testing.T) {
	mw := vary.New(vary.Options{Headers: []string{"Accept-Encoding"}})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "*")
		w.WriteHeader(http.StatusOK)
	})

	mw(inner).ServeHTTP(rec, req)

	// Vary: * means "varies on everything" — middleware should not append.
	if got := rec.Header().Get("Vary"); got != "*" {
		t.Errorf("expected Vary: * to remain unchanged, got %q", got)
	}
}
