package circuitbreaker_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/leodahal4/routeguard/circuitbreaker"
)

func newTestHandler(code int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := circuitbreaker.DefaultOptions()
	if opts.Threshold != 5 {
		t.Errorf("expected threshold 5, got %d", opts.Threshold)
	}
	if opts.Timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", opts.Timeout)
	}
	if opts.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", opts.StatusCode)
	}
}

func TestNew_SuccessfulRequests_CircuitStaysClosed(t *testing.T) {
	opts := circuitbreaker.DefaultOptions()
	mw := circuitbreaker.New(opts)(newTestHandler(http.StatusOK))

	for i := 0; i < 10; i++ {
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("iteration %d: expected 200, got %d", i, rec.Code)
		}
	}
}

func TestNew_FailuresOpenCircuit(t *testing.T) {
	opts := circuitbreaker.DefaultOptions()
	opts.Threshold = 3
	mw := circuitbreaker.New(opts)(newTestHandler(http.StatusInternalServerError))

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// Circuit should now be open.
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 when circuit open, got %d", rec.Code)
	}
}

func TestNew_CircuitHalfOpenAfterTimeout(t *testing.T) {
	opts := circuitbreaker.DefaultOptions()
	opts.Threshold = 1
	opts.Timeout = 50 * time.Millisecond
	mw := circuitbreaker.New(opts)(newTestHandler(http.StatusInternalServerError))

	// Trip the circuit.
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	// Immediately blocked.
	rec = httptest.NewRecorder()
	mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected circuit open, got %d", rec.Code)
	}

	time.Sleep(60 * time.Millisecond)

	// After timeout, half-open allows one request through.
	recSuccess := httptest.NewRecorder()
	successHandler := circuitbreaker.New(opts)(newTestHandler(http.StatusOK))
	successHandler.ServeHTTP(recSuccess, httptest.NewRequest(http.MethodGet, "/", nil))
	if recSuccess.Code != http.StatusOK {
		t.Errorf("expected request through in half-open, got %d", recSuccess.Code)
	}
}

func TestNew_SuccessResetsClosed(t *testing.T) {
	opts := circuitbreaker.DefaultOptions()
	opts.Threshold = 2
	opts.Timeout = 20 * time.Millisecond

	var serveCode int
	mw := circuitbreaker.New(opts)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(serveCode)
	}))

	serveCode = http.StatusInternalServerError
	for i := 0; i < 2; i++ {
		mw.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	}

	time.Sleep(30 * time.Millisecond)
	serveCode = http.StatusOK
	mw.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	// Circuit closed: next request should pass normally.
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("expected circuit closed after success, got %d", rec.Code)
	}
}
