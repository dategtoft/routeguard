package responsesize_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/patrickward/routeguard/responsesize"
)

func newTestHandler(body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := responsesize.DefaultOptions()
	if opts.MaxBytes != 10*1024*1024 {
		t.Errorf("expected MaxBytes=10MB, got %d", opts.MaxBytes)
	}
	if opts.ErrorMessage == "" {
		t.Error("expected non-empty default ErrorMessage")
	}
}

func TestNew_SmallResponse_Passes(t *testing.T) {
	handler := responsesize.New(responsesize.Options{
		MaxBytes: 100,
	})(newTestHandler("hello"))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "hello" {
		t.Errorf("expected body 'hello', got %q", rec.Body.String())
	}
}

func TestNew_LargeResponse_Returns500(t *testing.T) {
	body := strings.Repeat("x", 200)
	handler := responsesize.New(responsesize.Options{
		MaxBytes:     100,
		ErrorMessage: "too big",
	})(newTestHandler(body))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "too big") {
		t.Errorf("expected error message in body, got %q", rec.Body.String())
	}
}

func TestNew_ExactLimit_Passes(t *testing.T) {
	body := strings.Repeat("a", 50)
	handler := responsesize.New(responsesize.Options{
		MaxBytes: 50,
	})(newTestHandler(body))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestNew_DefaultOptions_Applied(t *testing.T) {
	// Zero-value MaxBytes should fall back to default (10MB).
	handler := responsesize.New(responsesize.Options{})(newTestHandler("small"))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
