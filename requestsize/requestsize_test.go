package requestsize_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/joeychilson/routeguard/requestsize"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := requestsize.DefaultOptions()
	if opts.MaxBytes != 1<<20 {
		t.Fatalf("expected MaxBytes=1048576, got %d", opts.MaxBytes)
	}
	if opts.ErrorMessage == "" {
		t.Fatal("expected non-empty ErrorMessage")
	}
	if len(opts.SkipMethods) == 0 {
		t.Fatal("expected default SkipMethods to be non-empty")
	}
}

func TestNew_SmallBody_Passes(t *testing.T) {
	h := requestsize.New(requestsize.DefaultOptions())(newTestHandler())
	body := strings.NewReader("hello")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestNew_ContentLengthExceeds_Returns413(t *testing.T) {
	opts := requestsize.DefaultOptions()
	opts.MaxBytes = 10
	h := requestsize.New(opts)(newTestHandler())

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("hello world!"))
	req.ContentLength = 12 // explicitly over limit
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rec.Code)
	}
}

func TestNew_StreamingBodyExceeds_Returns500(t *testing.T) {
	// When MaxBytesReader is exceeded during read the handler gets an error;
	// the default Go behaviour writes a 500 or the handler notices the error.
	// We verify the middleware at least wraps the body.
	opts := requestsize.DefaultOptions()
	opts.MaxBytes = 5

	h := requestsize.New(opts)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 100)
		_, err := r.Body.Read(buf)
		if err != nil {
			http.Error(w, "read error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	body := bytes.NewReader([]byte("this is definitely more than five bytes"))
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.ContentLength = -1 // unknown length — forces streaming path
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatal("expected non-200 when body exceeds limit")
	}
}

func TestNew_SkippedMethod_NotChecked(t *testing.T) {
	opts := requestsize.DefaultOptions()
	opts.MaxBytes = 1 // absurdly small
	h := requestsize.New(opts)(newTestHandler())

	// GET is in the default skip list — should pass regardless of body size.
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader("some body"))
	req.ContentLength = 9
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for skipped method, got %d", rec.Code)
	}
}
