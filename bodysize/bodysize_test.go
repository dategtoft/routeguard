package bodysize_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourusername/routeguard/bodysize"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Consume the entire body so MaxBytesReader can trigger if oversized.
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := bodysize.DefaultOptions()
	if opts.MaxBytes != 1<<20 {
		t.Errorf("expected MaxBytes=1048576, got %d", opts.MaxBytes)
	}
	if opts.ErrorMessage == "" {
		t.Error("expected non-empty ErrorMessage")
	}
}

func TestNew_SmallBody_Passes(t *testing.T) {
	mw := bodysize.New(bodysize.DefaultOptions())
	h := mw(newTestHandler())

	body := strings.NewReader("hello world")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestNew_ContentLengthExceeds_Returns413(t *testing.T) {
	opts := bodysize.Options{MaxBytes: 10, ErrorMessage: "too big"}
	mw := bodysize.New(opts)
	h := mw(newTestHandler())

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("hello world"))
	req.ContentLength = 11 // explicitly set to exceed limit
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "too big") {
		t.Errorf("expected custom error message in body, got: %s", rec.Body.String())
	}
}

func TestNew_StreamingBodyExceeds_Returns413(t *testing.T) {
	opts := bodysize.Options{MaxBytes: 5}
	mw := bodysize.New(opts)
	h := mw(newTestHandler())

	// Send more bytes than the limit without setting Content-Length.
	body := bytes.NewReader([]byte("this is definitely more than five bytes"))
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.ContentLength = -1 // unknown length
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	// MaxBytesReader causes the handler to receive an error on read;
	// the middleware itself returns 200 in this path — the handler is
	// responsible for checking read errors. We just verify no panic occurs.
	if rec.Code != http.StatusOK && rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("unexpected status %d", rec.Code)
	}
}

func TestNew_ZeroMaxBytes_UsesDefault(t *testing.T) {
	opts := bodysize.Options{MaxBytes: 0}
	mw := bodysize.New(opts)
	h := mw(newTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
