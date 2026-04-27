package timeout_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/routeguard/timeout"
)

func newSlowHandler(delay time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(delay):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok")) //nolint:errcheck
		case <-r.Context().Done():
		}
	})
}

func newFastHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok")) //nolint:errcheck
	})
}

func TestTimeout_RequestCompletesInTime(t *testing.T) {
	mw := timeout.New(timeout.Options{Duration: 500 * time.Millisecond})
	handler := mw(newFastHandler())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if rr.Body.String() != "ok" {
		t.Errorf("expected body 'ok', got %q", rr.Body.String())
	}
}

func TestTimeout_RequestExceedsTimeout(t *testing.T) {
	mw := timeout.New(timeout.Options{
		Duration:   50 * time.Millisecond,
		Message:    "request timed out",
		StatusCode: http.StatusGatewayTimeout,
	})
	handler := mw(newSlowHandler(200 * time.Millisecond))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusGatewayTimeout {
		t.Errorf("expected 504, got %d", rr.Code)
	}
	if rr.Body.String() != "request timed out" {
		t.Errorf("expected timeout message, got %q", rr.Body.String())
	}
}

func TestTimeout_DefaultOptions(t *testing.T) {
	opts := timeout.DefaultOptions()
	if opts.Duration != 30*time.Second {
		t.Errorf("expected 30s, got %v", opts.Duration)
	}
	if opts.StatusCode != http.StatusGatewayTimeout {
		t.Errorf("expected 504, got %d", opts.StatusCode)
	}
	if opts.Message == "" {
		t.Error("expected non-empty default message")
	}
}

func TestTimeout_ZeroDurationUsesDefault(t *testing.T) {
	// Should not panic and should use default duration
	mw := timeout.New(timeout.Options{Duration: 0})
	handler := mw(newFastHandler())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestTimeout_CustomStatusCode(t *testing.T) {
	mw := timeout.New(timeout.Options{
		Duration:   30 * time.Millisecond,
		StatusCode: http.StatusRequestTimeout,
		Message:    "too slow",
	})
	handler := mw(newSlowHandler(200 * time.Millisecond))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestTimeout {
		t.Errorf("expected 408, got %d", rr.Code)
	}
	if rr.Body.String() != "too slow" {
		t.Errorf("expected 'too slow', got %q", rr.Body.String())
	}
}
