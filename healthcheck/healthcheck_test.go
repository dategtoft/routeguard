package healthcheck_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/routeguard/healthcheck"
)

func TestNew_DefaultOptions_Returns200(t *testing.T) {
	h := healthcheck.New(healthcheck.DefaultOptions())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp healthcheck.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != healthcheck.StatusOK {
		t.Errorf("expected status ok, got %s", resp.Status)
	}
}

func TestNew_AllChecksPass_Returns200(t *testing.T) {
	opts := healthcheck.DefaultOptions()
	opts.Checks["db"] = func() error { return nil }
	opts.Checks["cache"] = func() error { return nil }

	h := healthcheck.New(opts)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp healthcheck.Response
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Checks["db"] != "ok" || resp.Checks["cache"] != "ok" {
		t.Errorf("unexpected check values: %v", resp.Checks)
	}
}

func TestNew_FailingCheck_Returns503(t *testing.T) {
	opts := healthcheck.DefaultOptions()
	opts.Checks["db"] = func() error { return errors.New("connection refused") }

	h := healthcheck.New(opts)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}

	var resp healthcheck.Response
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Status != healthcheck.StatusDegraded {
		t.Errorf("expected degraded, got %s", resp.Status)
	}
	if resp.Checks["db"] != "connection refused" {
		t.Errorf("unexpected check message: %s", resp.Checks["db"])
	}
}

func TestNew_ContentTypeIsJSON(t *testing.T) {
	h := healthcheck.New(healthcheck.DefaultOptions())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}

func TestNew_TimestampPresent(t *testing.T) {
	h := healthcheck.New(healthcheck.DefaultOptions())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	var resp healthcheck.Response
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}
