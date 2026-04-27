package recovery_test

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourusername/routeguard/recovery"
)

func newPanicHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	})
}

func newOKHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

func TestRecovery_NoPanic(t *testing.T) {
	handler := recovery.New(newOKHandler(), recovery.DefaultOptions())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestRecovery_PanicReturns500(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	opts := recovery.Options{Logger: logger, EnableStackTrace: false}

	handler := recovery.New(newPanicHandler(), opts)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestRecovery_LogsPanicMessage(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	opts := recovery.Options{Logger: logger, EnableStackTrace: false}

	handler := recovery.New(newPanicHandler(), opts)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(rec, req)

	if !strings.Contains(buf.String(), "something went wrong") {
		t.Errorf("expected panic message in log, got: %s", buf.String())
	}
}

func TestRecovery_LogsStackTrace(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	opts := recovery.Options{Logger: logger, EnableStackTrace: true}

	handler := recovery.New(newPanicHandler(), opts)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(rec, req)

	if !strings.Contains(buf.String(), "goroutine") {
		t.Errorf("expected stack trace in log, got: %s", buf.String())
	}
}

func TestRecovery_DefaultOptions(t *testing.T) {
	handler := recovery.New(newPanicHandler(), recovery.DefaultOptions())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Should not panic itself
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}
