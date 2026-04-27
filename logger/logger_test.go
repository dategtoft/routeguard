package logger

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestHandler(code int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	})
}

func TestNew_DefaultPrefix(t *testing.T) {
	l := New("")
	if l.prefix != "[routeguard]" {
		t.Errorf("expected default prefix '[routeguard]', got %q", l.prefix)
	}
}

func TestNew_CustomPrefix(t *testing.T) {
	l := New("[test]")
	if l.prefix != "[test]" {
		t.Errorf("expected prefix '[test]', got %q", l.prefix)
	}
}

func TestMiddleware_LogsRequest(t *testing.T) {
	var buf bytes.Buffer
	l := New("[test]")
	l.log = log.New(&buf, "[test] ", 0)

	handler := l.Middleware(newTestHandler(http.StatusOK))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	output := buf.String()
	if !strings.Contains(output, "GET") {
		t.Errorf("expected log to contain method 'GET', got: %s", output)
	}
	if !strings.Contains(output, "/health") {
		t.Errorf("expected log to contain path '/health', got: %s", output)
	}
	if !strings.Contains(output, "200") {
		t.Errorf("expected log to contain status '200', got: %s", output)
	}
}

func TestMiddleware_LogsNonOKStatus(t *testing.T) {
	var buf bytes.Buffer
	l := New("[test]")
	l.log = log.New(&buf, "[test] ", 0)

	handler := l.Middleware(newTestHandler(http.StatusNotFound))

	req := httptest.NewRequest(http.MethodPost, "/missing", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	output := buf.String()
	if !strings.Contains(output, "404") {
		t.Errorf("expected log to contain status '404', got: %s", output)
	}
}
