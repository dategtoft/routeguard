package throttle_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/patrickward/routeguard/throttle"
)

func newTestHandler(delay time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if delay > 0 {
			time.Sleep(delay)
		}
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := throttle.DefaultOptions()
	if opts.MaxConcurrent != 100 {
		t.Errorf("expected MaxConcurrent=100, got %d", opts.MaxConcurrent)
	}
	if opts.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected StatusCode=503, got %d", opts.StatusCode)
	}
}

func TestNew_AllowsUnderLimit(t *testing.T) {
	opts := throttle.DefaultOptions()
	opts.MaxConcurrent = 5
	mw := throttle.New(opts)
	handler := mw(newTestHandler(0))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestNew_BlocksWhenFull(t *testing.T) {
	opts := throttle.Options{
		MaxConcurrent: 1,
		MaxQueueSize:  0,
		Timeout:       100 * time.Millisecond,
		StatusCode:    http.StatusServiceUnavailable,
		Message:       "busy",
	}
	mw := throttle.New(opts)

	// Hold the single slot open.
	blocking := make(chan struct{})
	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-blocking
		w.WriteHeader(http.StatusOK)
	})
	handler := mw(slowHandler)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		handler.ServeHTTP(rr, req)
	}()

	time.Sleep(20 * time.Millisecond) // let goroutine acquire slot

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr.Code)
	}

	close(blocking)
	wg.Wait()
}

func TestNew_QueuedRequestEventuallyServed(t *testing.T) {
	opts := throttle.Options{
		MaxConcurrent: 1,
		MaxQueueSize:  5,
		Timeout:       2 * time.Second,
		StatusCode:    http.StatusServiceUnavailable,
		Message:       "busy",
	}
	mw := throttle.New(opts)

	blocking := make(chan struct{})
	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-blocking
		w.WriteHeader(http.StatusOK)
	})
	handler := mw(slowHandler)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		handler.ServeHTTP(rr, req)
	}()

	time.Sleep(20 * time.Millisecond)

	resultCh := make(chan int, 1)
	go func() {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		handler.ServeHTTP(rr, req)
		resultCh <- rr.Code
	}()

	time.Sleep(20 * time.Millisecond)
	close(blocking)
	wg.Wait()

	code := <-resultCh
	if code != http.StatusOK {
		t.Errorf("expected queued request to succeed with 200, got %d", code)
	}
}
