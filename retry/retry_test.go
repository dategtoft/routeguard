package retry_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/patrickward/routeguard/retry"
)

func newCountingHandler(code int, callCount *int32) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(callCount, 1)
		w.WriteHeader(code)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := retry.DefaultOptions()
	if opts.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts=3, got %d", opts.MaxAttempts)
	}
	if opts.Delay != 100*time.Millisecond {
		t.Errorf("expected Delay=100ms, got %v", opts.Delay)
	}
	if opts.ShouldRetry == nil {
		t.Error("expected ShouldRetry to be non-nil")
	}
}

func TestNew_SuccessOnFirstAttempt(t *testing.T) {
	var calls int32
	h := retry.New(retry.DefaultOptions())(newCountingHandler(http.StatusOK, &calls))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestNew_RetriesOnServerError(t *testing.T) {
	var calls int32
	opts := retry.DefaultOptions()
	opts.Delay = 0
	h := retry.New(opts)(newCountingHandler(http.StatusServiceUnavailable, &calls))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if atomic.LoadInt32(&calls) != int32(opts.MaxAttempts) {
		t.Errorf("expected %d calls, got %d", opts.MaxAttempts, calls)
	}
}

func TestNew_NoRetryOn404(t *testing.T) {
	var calls int32
	opts := retry.DefaultOptions()
	opts.Delay = 0
	h := retry.New(opts)(newCountingHandler(http.StatusNotFound, &calls))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestNew_CustomShouldRetry(t *testing.T) {
	var calls int32
	opts := retry.Options{
		MaxAttempts: 2,
		Delay:       0,
		ShouldRetry: func(code int) bool { return code == http.StatusNotFound },
	}
	h := retry.New(opts)(newCountingHandler(http.StatusNotFound, &calls))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if atomic.LoadInt32(&calls) != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

func TestNew_EventualSuccess(t *testing.T) {
	var calls int32
	// Fails twice, succeeds on third attempt.
	h := retry.New(retry.Options{
		MaxAttempts: 3,
		Delay:       0,
		ShouldRetry: func(code int) bool { return code >= 500 },
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if atomic.LoadInt32(&calls) != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}
