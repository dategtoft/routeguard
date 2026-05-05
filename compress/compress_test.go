package compress_test

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourusername/routeguard/compress"
)

func newTestHandler(body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := compress.DefaultOptions()
	if opts.Level != gzip.DefaultCompression {
		t.Errorf("expected default compression level, got %d", opts.Level)
	}
	if opts.MinLength != 1024 {
		t.Errorf("expected MinLength 1024, got %d", opts.MinLength)
	}
}

func TestNew_CompressesWhenAccepted(t *testing.T) {
	body := strings.Repeat("hello routeguard ", 100)
	mw := compress.New(compress.DefaultOptions())
	handler := mw(newTestHandler(body))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatal("expected Content-Encoding: gzip")
	}

	gr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gr.Close()

	decompressed, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("failed to decompress: %v", err)
	}
	if string(decompressed) != body {
		t.Errorf("decompressed body mismatch")
	}
}

func TestNew_NoCompressionWithoutHeader(t *testing.T) {
	body := "plain response"
	mw := compress.New(compress.DefaultOptions())
	handler := mw(newTestHandler(body))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") == "gzip" {
		t.Fatal("expected no gzip encoding")
	}
	if rec.Body.String() != body {
		t.Errorf("expected plain body %q, got %q", body, rec.Body.String())
	}
}

func TestNew_ContentLengthRemoved(t *testing.T) {
	mw := compress.New(compress.DefaultOptions())
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "999")
		_, _ = w.Write([]byte("data"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Length") != "" {
		t.Error("expected Content-Length to be removed when gzip is applied")
	}
}

func TestNew_BelowMinLengthNotCompressed(t *testing.T) {
	// Body shorter than MinLength (1024) should not be compressed even if
	// the client advertises gzip support.
	body := "short"
	mw := compress.New(compress.DefaultOptions())
	handler := mw(newTestHandler(body))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") == "gzip" {
		t.Error("expected no gzip encoding for body below MinLength")
	}
	if rec.Body.String() != body {
		t.Errorf("expected plain body %q, got %q", body, rec.Body.String())
	}
}
