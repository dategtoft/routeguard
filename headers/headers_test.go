package headers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/routeguard/headers"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := headers.DefaultOptions()
	if !opts.SecurityHeaders {
		t.Error("expected SecurityHeaders to be true")
	}
	if opts.HSTSMaxAge != 31536000 {
		t.Errorf("expected HSTSMaxAge 31536000, got %d", opts.HSTSMaxAge)
	}
}

func TestNew_SecurityHeaders_Set(t *testing.T) {
	h := headers.New(headers.DefaultOptions())(newTestHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	expected := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-Xss-Protection":       "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
		"Content-Security-Policy": "default-src 'self'",
	}
	for k, want := range expected {
		if got := rec.Header().Get(k); got != want {
			t.Errorf("header %s: got %q, want %q", k, got, want)
		}
	}
}

func TestNew_HSTSHeader(t *testing.T) {
	h := headers.New(headers.Options{SecurityHeaders: true, HSTSMaxAge: 3600})(newTestHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	want := "max-age=3600; includeSubDomains"
	if got := rec.Header().Get("Strict-Transport-Security"); got != want {
		t.Errorf("HSTS: got %q, want %q", got, want)
	}
}

func TestNew_SecurityHeaders_Disabled(t *testing.T) {
	h := headers.New(headers.Options{SecurityHeaders: false})(newTestHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Header().Get("X-Frame-Options"); got != "" {
		t.Errorf("expected no X-Frame-Options, got %q", got)
	}
}

func TestNew_CustomHeaders(t *testing.T) {
	opts := headers.Options{
		Custom: map[string]string{
			"X-App-Version": "1.2.3",
			"X-Request-Source": "api",
		},
	}
	h := headers.New(opts)(newTestHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Header().Get("X-App-Version"); got != "1.2.3" {
		t.Errorf("X-App-Version: got %q, want %q", got, "1.2.3")
	}
	if got := rec.Header().Get("X-Request-Source"); got != "api" {
		t.Errorf("X-Request-Source: got %q, want %q", got, "api")
	}
}

func TestNew_PassThrough(t *testing.T) {
	h := headers.New(headers.DefaultOptions())(newTestHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
