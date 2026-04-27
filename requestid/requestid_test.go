package requestid_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/routeguard/requestid"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestNew_GeneratesID(t *testing.T) {
	mw := requestid.New(requestid.DefaultOptions())
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	mw(newTestHandler()).ServeHTTP(rr, req)

	id := rr.Header().Get(requestid.DefaultHeader)
	if id == "" {
		t.Fatal("expected a request ID in response header, got empty string")
	}
}

func TestNew_ReusesIncomingID(t *testing.T) {
	const existingID = "my-custom-id-123"

	mw := requestid.New(requestid.DefaultOptions())
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(requestid.DefaultHeader, existingID)

	mw(newTestHandler()).ServeHTTP(rr, req)

	if got := rr.Header().Get(requestid.DefaultHeader); got != existingID {
		t.Fatalf("expected %q, got %q", existingID, got)
	}
}

func TestNew_IDStoredInContext(t *testing.T) {
	var capturedID string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = requestid.FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	mw := requestid.New(requestid.DefaultOptions())
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	mw(handler).ServeHTTP(rr, req)

	if capturedID == "" {
		t.Fatal("expected request ID in context, got empty string")
	}
	if capturedID != rr.Header().Get(requestid.DefaultHeader) {
		t.Fatalf("context ID %q does not match header ID %q", capturedID, rr.Header().Get(requestid.DefaultHeader))
	}
}

func TestNew_CustomHeader(t *testing.T) {
	const customHeader = "X-Trace-ID"

	opts := requestid.Options{
		Header: customHeader,
	}
	mw := requestid.New(opts)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	mw(newTestHandler()).ServeHTTP(rr, req)

	if id := rr.Header().Get(customHeader); id == "" {
		t.Fatalf("expected request ID in %q header, got empty string", customHeader)
	}
}

func TestNew_CustomGenerator(t *testing.T) {
	const fixedID = "fixed-id"

	opts := requestid.Options{
		Generator: func() string { return fixedID },
	}
	mw := requestid.New(opts)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	mw(newTestHandler()).ServeHTTP(rr, req)

	if got := rr.Header().Get(requestid.DefaultHeader); got != fixedID {
		t.Fatalf("expected %q, got %q", fixedID, got)
	}
}

func TestFromContext_Empty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if id := requestid.FromContext(req.Context()); id != "" {
		t.Fatalf("expected empty string, got %q", id)
	}
}
