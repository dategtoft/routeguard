package sanitize_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/joeychilson/routeguard/sanitize"
)

func newCaptureHandler(captured *url.Values) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*captured = r.URL.Query()
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := sanitize.DefaultOptions()
	if !opts.EscapeHTML {
		t.Error("expected EscapeHTML to be true")
	}
	if !opts.StripNullBytes {
		t.Error("expected StripNullBytes to be true")
	}
	if opts.TrimSpace {
		t.Error("expected TrimSpace to be false")
	}
}

func TestNew_PassThrough_CleanInput(t *testing.T) {
	var captured url.Values
	h := sanitize.New(sanitize.DefaultOptions())(newCaptureHandler(&captured))

	req := httptest.NewRequest(http.MethodGet, "/?name=John", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := captured.Get("name"); got != "John" {
		t.Errorf("expected John, got %s", got)
	}
}

func TestNew_EscapesHTML(t *testing.T) {
	var captured url.Values
	h := sanitize.New(sanitize.DefaultOptions())(newCaptureHandler(&captured))

	req := httptest.NewRequest(http.MethodGet, "/?q=%3Cscript%3Ealert(1)%3C%2Fscript%3E", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	got := captured.Get("q")
	if got == "<script>alert(1)</script>" {
		t.Error("HTML was not escaped")
	}
	if got != "&lt;script&gt;alert(1)&lt;/script&gt;" {
		t.Errorf("unexpected escaped value: %s", got)
	}
}

func TestNew_StripNullBytes(t *testing.T) {
	var captured url.Values
	h := sanitize.New(sanitize.DefaultOptions())(newCaptureHandler(&captured))

	req := httptest.NewRequest(http.MethodGet, "/?data=hel%00lo", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := captured.Get("data"); got != "hello" {
		t.Errorf("expected null bytes stripped, got %q", got)
	}
}

func TestNew_TrimSpace(t *testing.T) {
	var captured url.Values
	opts := sanitize.DefaultOptions()
	opts.TrimSpace = true
	h := sanitize.New(opts)(newCaptureHandler(&captured))

	req := httptest.NewRequest(http.MethodGet, "/?name=+hello+", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := captured.Get("name"); got != "hello" {
		t.Errorf("expected trimmed value, got %q", got)
	}
}

func TestNew_NoEscapeHTML_WhenDisabled(t *testing.T) {
	var captured url.Values
	opts := sanitize.DefaultOptions()
	opts.EscapeHTML = false
	h := sanitize.New(opts)(newCaptureHandler(&captured))

	req := httptest.NewRequest(http.MethodGet, "/?q=%3Cb%3E", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := captured.Get("q"); got != "<b>" {
		t.Errorf("expected unescaped value, got %q", got)
	}
}
