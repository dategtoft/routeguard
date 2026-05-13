package observe_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrickward/routeguard/observe"
)

func newTestHandler(status int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := observe.DefaultOptions()
	if opts.Path != "/metrics" {
		t.Fatalf("expected /metrics, got %s", opts.Path)
	}
}

func TestNew_PassThrough(t *testing.T) {
	h := observe.New(newTestHandler(http.StatusOK), observe.DefaultOptions())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/hello", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestNew_MetricsEndpoint_Returns200(t *testing.T) {
	h := observe.New(newTestHandler(http.StatusOK), observe.DefaultOptions())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
}

func TestNew_CountsRequests(t *testing.T) {
	h := observe.New(newTestHandler(http.StatusOK), observe.DefaultOptions())
	for i := 0; i < 3; i++ {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/ping", nil))
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	var payload map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if payload["requests_total"].(float64) != 3 {
		t.Fatalf("expected 3 requests, got %v", payload["requests_total"])
	}
}

func TestNew_CountsErrors(t *testing.T) {
	h := observe.New(newTestHandler(http.StatusInternalServerError), observe.DefaultOptions())
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/boom", nil))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	var payload map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if payload["errors_total"].(float64) != 1 {
		t.Fatalf("expected 1 error, got %v", payload["errors_total"])
	}
}

func TestNew_Namespace(t *testing.T) {
	opts := observe.Options{Path: "/metrics", Namespace: "app"}
	h := observe.New(newTestHandler(http.StatusOK), opts)
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/x", nil))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	var payload map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if _, ok := payload["app_requests_total"]; !ok {
		t.Fatalf("expected key app_requests_total in %v", payload)
	}
}
