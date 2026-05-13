package reqid_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrickward/routeguard/reqid"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := reqid.DefaultOptions()
	if opts.Header != "X-Request-ID" {
		t.Errorf("expected header X-Request-ID, got %s", opts.Header)
	}
	if opts.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", opts.StatusCode)
	}
	if opts.Message == "" {
		t.Error("expected non-empty default message")
	}
}

func TestNew_MissingHeader_Returns400(t *testing.T) {
	mw := reqid.New(reqid.DefaultOptions())(newTestHandler())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestNew_PresentHeader_PassesThrough(t *testing.T) {
	mw := reqid.New(reqid.DefaultOptions())(newTestHandler())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "abc-123")
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestNew_CustomHeader_Enforced(t *testing.T) {
	opts := reqid.Options{
		Header:     "X-Correlation-ID",
		Message:    "correlation ID required",
		StatusCode: http.StatusUnprocessableEntity,
	}
	mw := reqid.New(opts)(newTestHandler())

	// missing custom header
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}

	// present custom header
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("X-Correlation-ID", "xyz-789")
	mw.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr2.Code)
	}
}

func TestNew_DefaultsAppliedWhenZeroValue(t *testing.T) {
	mw := reqid.New(reqid.Options{})(newTestHandler())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}
