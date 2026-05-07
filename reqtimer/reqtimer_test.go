package reqtimer_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/joeydtaylor/routeguard/reqtimer"
)

func newTestHandler(sleep time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if sleep > 0 {
			time.Sleep(sleep)
		}
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := reqtimer.DefaultOptions()
	if opts.Header != "X-Request-Duration" {
		t.Fatalf("expected default header X-Request-Duration, got %s", opts.Header)
	}
	if opts.Precision != time.Millisecond {
		t.Fatalf("expected default precision Millisecond, got %v", opts.Precision)
	}
}

func TestNew_HeaderPresent(t *testing.T) {
	mw := reqtimer.New(reqtimer.DefaultOptions())
	h := mw(newTestHandler(0))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	val := rec.Header().Get("X-Request-Duration")
	if val == "" {
		t.Fatal("expected X-Request-Duration header to be set")
	}
}

func TestNew_DurationIsNonNegative(t *testing.T) {
	mw := reqtimer.New(reqtimer.DefaultOptions())
	h := mw(newTestHandler(10 * time.Millisecond))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	val := rec.Header().Get("X-Request-Duration")
	raw := strings.TrimSuffix(val, "ms")
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		t.Fatalf("could not parse duration value %q: %v", val, err)
	}
	if f < 0 {
		t.Fatalf("expected non-negative duration, got %f", f)
	}
}

func TestNew_CustomHeader(t *testing.T) {
	opts := reqtimer.Options{
		Header:    "X-Elapsed",
		Precision: time.Millisecond,
	}
	mw := reqtimer.New(opts)
	h := mw(newTestHandler(0))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Header().Get("X-Elapsed") == "" {
		t.Fatal("expected X-Elapsed header to be set")
	}
	if rec.Header().Get("X-Request-Duration") != "" {
		t.Fatal("default header should not be set when custom header is used")
	}
}

func TestNew_SecondPrecision(t *testing.T) {
	opts := reqtimer.Options{
		Header:    "X-Request-Duration",
		Precision: time.Second,
	}
	mw := reqtimer.New(opts)
	h := mw(newTestHandler(0))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	val := rec.Header().Get("X-Request-Duration")
	if !strings.HasSuffix(val, "s") {
		t.Fatalf("expected value to end with 's', got %q", val)
	}
}
