package dedupe_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/patrickward/routeguard/dedupe"
)

func newCountingHandler(delay time.Duration) (http.Handler, *atomic.Int64) {
	var calls atomic.Int64
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if delay > 0 {
			time.Sleep(delay)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	})
	return h, &calls
}

func TestDefaultOptions(t *testing.T) {
	opts := dedupe.DefaultOptions()
	if opts.KeyFunc == nil {
		t.Fatal("expected non-nil KeyFunc")
	}
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	key := opts.KeyFunc(r)
	if key != "GET:/test" {
		t.Fatalf("expected 'GET:/test', got %q", key)
	}
}

func TestNew_SingleRequest_Passes(t *testing.T) {
	h, calls := newCountingHandler(0)
	mw := dedupe.New(dedupe.DefaultOptions())(h)

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "hello" {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
	if calls.Load() != 1 {
		t.Fatalf("expected 1 upstream call, got %d", calls.Load())
	}
}

func TestNew_ConcurrentDuplicates_CoalescedIntoOne(t *testing.T) {
	h, calls := newCountingHandler(40 * time.Millisecond)
	mw := dedupe.New(dedupe.DefaultOptions())(h)

	const n = 10
	var wg sync.WaitGroup
	results := make([]int, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			rec := httptest.NewRecorder()
			mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/resource", nil))
			results[idx] = rec.Code
		}(i)
	}
	wg.Wait()

	for i, code := range results {
		if code != http.StatusOK {
			t.Errorf("goroutine %d: expected 200, got %d", i, code)
		}
	}
	if calls.Load() != 1 {
		t.Fatalf("expected exactly 1 upstream call, got %d", calls.Load())
	}
}

func TestNew_DifferentKeys_NotCoalesced(t *testing.T) {
	h, calls := newCountingHandler(20 * time.Millisecond)
	mw := dedupe.New(dedupe.DefaultOptions())(h)

	var wg sync.WaitGroup
	for _, path := range []string{"/a", "/b", "/c"} {
		wg.Add(1)
		p := path
		go func() {
			defer wg.Done()
			rec := httptest.NewRecorder()
			mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, p, nil))
		}()
	}
	wg.Wait()

	if calls.Load() != 3 {
		t.Fatalf("expected 3 upstream calls, got %d", calls.Load())
	}
}

func TestNew_SequentialRequests_NotCoalesced(t *testing.T) {
	h, calls := newCountingHandler(0)
	mw := dedupe.New(dedupe.DefaultOptions())(h)

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/seq", nil))
	}
	if calls.Load() != 3 {
		t.Fatalf("expected 3 upstream calls, got %d", calls.Load())
	}
}

func TestNew_CustomKeyFunc(t *testing.T) {
	h, calls := newCountingHandler(20 * time.Millisecond)
	opts := dedupe.Options{
		KeyFunc: func(r *http.Request) string { return "static-key" },
	}
	mw := dedupe.New(opts)(h)

	var wg sync.WaitGroup
	for _, path := range []string{"/x", "/y"} {
		wg.Add(1)
		p := path
		go func() {
			defer wg.Done()
			rec := httptest.NewRecorder()
			mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, p, nil))
		}()
	}
	wg.Wait()

	if calls.Load() != 1 {
		t.Fatalf("expected 1 upstream call with static key, got %d", calls.Load())
	}
}
