package maxconns_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/yourusername/routeguard/maxconns"
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
	opts := maxconns.DefaultOptions()
	if opts.Max != 100 {
		t.Errorf("expected Max=100, got %d", opts.Max)
	}
	if opts.Message == "" {
		t.Error("expected non-empty Message")
	}
}

func TestNew_AllowsUnderLimit(t *testing.T) {
	h := maxconns.New(newTestHandler(0), maxconns.Options{Max: 5})

	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if rec.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i, rec.Code)
		}
	}
}

func TestNew_BlocksWhenLimitExceeded(t *testing.T) {
	const limit = 3
	blocked := make(chan struct{})

	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-blocked
		w.WriteHeader(http.StatusOK)
	})

	h := maxconns.New(slow, maxconns.Options{Max: limit})

	var wg sync.WaitGroup
	results := make([]int, limit+2)

	for i := 0; i < limit+2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
			results[idx] = rec.Code
		}(i)
	}

	// give goroutines time to reach the handler
	time.Sleep(20 * time.Millisecond)
	close(blocked)
	wg.Wait()

	rejected := 0
	for _, code := range results {
		if code == http.StatusServiceUnavailable {
			rejected++
		}
	}
	if rejected < 1 {
		t.Errorf("expected at least 1 rejected request, got %d", rejected)
	}
}

func TestNew_CustomMessage(t *testing.T) {
	h := maxconns.New(newTestHandler(time.Second), maxconns.Options{
		Max:     1,
		Message: "too busy",
	})

	blocked := make(chan struct{})
	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-blocked
		w.WriteHeader(http.StatusOK)
	})
	h = maxconns.New(slow, maxconns.Options{Max: 1, Message: "too busy"})

	var wg sync.WaitGroup
	results := make(chan *httptest.ResponseRecorder, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
			results <- rec
		}()
	}

	time.Sleep(20 * time.Millisecond)
	close(blocked)
	wg.Wait()
	close(results)

	for rec := range results {
		if rec.Code == http.StatusServiceUnavailable {
			if body := rec.Body.String(); body != "too busy\n" {
				t.Errorf("expected body 'too busy\\n', got %q", body)
			}
		}
	}
}

func TestNew_ZeroValueUsesDefaults(t *testing.T) {
	h := maxconns.New(newTestHandler(0), maxconns.Options{})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
