package requestlog_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourusername/routeguard/requestlog"
)

func newTestHandler(status int, body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := requestlog.DefaultOptions()
	if opts.Writer == nil {
		t.Fatal("expected non-nil Writer")
	}
	if opts.TimeFormat == "" {
		t.Fatal("expected non-empty TimeFormat")
	}
}

func TestNew_LogsRequest(t *testing.T) {
	var buf bytes.Buffer
	opts := requestlog.DefaultOptions()
	opts.Writer = &buf

	mw := requestlog.New(opts)
	handler := mw(newTestHandler(http.StatusOK, "hello"))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	log := buf.String()
	for _, want := range []string{"method=GET", "path=/test", "status=200", "bytes=5"} {
		if !strings.Contains(log, want) {
			t.Errorf("log missing %q; got: %s", want, log)
		}
	}
}

func TestNew_LogsNon200Status(t *testing.T) {
	var buf bytes.Buffer
	opts := requestlog.DefaultOptions()
	opts.Writer = &buf

	mw := requestlog.New(opts)
	handler := mw(newTestHandler(http.StatusNotFound, "not found"))

	req := httptest.NewRequest(http.MethodPost, "/missing", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	log := buf.String()
	if !strings.Contains(log, "status=404") {
		t.Errorf("expected status=404 in log; got: %s", log)
	}
}

func TestNew_SkipPath(t *testing.T) {
	var buf bytes.Buffer
	opts := requestlog.DefaultOptions()
	opts.Writer = &buf
	opts.SkipPaths = []string{"/healthz"}

	mw := requestlog.New(opts)
	handler := mw(newTestHandler(http.StatusOK, "ok"))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if buf.Len() != 0 {
		t.Errorf("expected no log output for skipped path; got: %s", buf.String())
	}
}

func TestNew_DurationPresent(t *testing.T) {
	var buf bytes.Buffer
	opts := requestlog.DefaultOptions()
	opts.Writer = &buf

	mw := requestlog.New(opts)
	handler := mw(newTestHandler(http.StatusOK, ""))

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if !strings.Contains(buf.String(), "duration=") {
		t.Errorf("expected duration field in log; got: %s", buf.String())
	}
}
