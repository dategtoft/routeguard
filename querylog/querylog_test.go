package querylog_test

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/patrickward/routeguard/querylog"
)

func newTestHandler(code int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := querylog.DefaultOptions()
	if opts.Logger == nil {
		t.Fatal("expected non-nil logger")
	}
	if opts.Prefix == "" {
		t.Fatal("expected non-empty prefix")
	}
}

func TestNew_NoQueryParams_NoLog(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	opts := querylog.DefaultOptions()
	opts.Logger = logger

	mw := querylog.New(opts)(newTestHandler(http.StatusOK))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/path", nil)
	mw.ServeHTTP(rec, req)

	if buf.Len() != 0 {
		t.Errorf("expected no log output, got: %s", buf.String())
	}
}

func TestNew_LogsQueryParams(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	opts := querylog.DefaultOptions()
	opts.Logger = logger

	mw := querylog.New(opts)(newTestHandler(http.StatusOK))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/search?q=hello&page=2", nil)
	mw.ServeHTTP(rec, req)

	got := buf.String()
	if !strings.Contains(got, "q=hello") && !strings.Contains(got, "page=2") {
		t.Errorf("expected query params in log, got: %s", got)
	}
}

func TestNew_RedactsKeys(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	opts := querylog.DefaultOptions()
	opts.Logger = logger
	opts.RedactKeys = []string{"token", "secret"}

	mw := querylog.New(opts)(newTestHandler(http.StatusOK))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api?token=abc123&user=alice", nil)
	mw.ServeHTTP(rec, req)

	got := buf.String()
	if strings.Contains(got, "abc123") {
		t.Errorf("expected token value to be redacted, got: %s", got)
	}
	if !strings.Contains(got, "[REDACTED]") {
		t.Errorf("expected [REDACTED] in log, got: %s", got)
	}
	if !strings.Contains(got, "user=alice") {
		t.Errorf("expected user=alice in log, got: %s", got)
	}
}

func TestNew_SkipPaths(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	opts := querylog.DefaultOptions()
	opts.Logger = logger
	opts.SkipPaths = []string{"/health"}

	mw := querylog.New(opts)(newTestHandler(http.StatusOK))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health?check=true", nil)
	mw.ServeHTTP(rec, req)

	if buf.Len() != 0 {
		t.Errorf("expected no log for skipped path, got: %s", buf.String())
	}
}
