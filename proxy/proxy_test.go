package proxy_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/patrickward/routeguard/proxy"
)

func newBackend(body string, status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

func TestDefaultOptions(t *testing.T) {
	opts := proxy.DefaultOptions("http://example.com")
	if opts.Target != "http://example.com" {
		t.Fatalf("expected target http://example.com, got %s", opts.Target)
	}
	if opts.Timeout != 30*time.Second {
		t.Fatalf("expected 30s timeout, got %v", opts.Timeout)
	}
}

func TestNew_ProxiesRequest(t *testing.T) {
	backend := newBackend("hello from backend", http.StatusOK)
	defer backend.Close()

	mw := proxy.New(proxy.DefaultOptions(backend.URL))
	handler := mw(http.NotFoundHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "hello from backend" {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestNew_StripPrefix(t *testing.T) {
	var receivedPath string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	opts := proxy.DefaultOptions(backend.URL)
	opts.StripPrefix = "/api"
	mw := proxy.New(opts)
	handler := mw(http.NotFoundHandler())

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if receivedPath != "/users" {
		t.Fatalf("expected /users, got %s", receivedPath)
	}
}

func TestNew_ForwardsXForwardedHost(t *testing.T) {
	var receivedHeader string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header.Get("X-Forwarded-Host")
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	mw := proxy.New(proxy.DefaultOptions(backend.URL))
	handler := mw(http.NotFoundHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "myapp.example.com"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if receivedHeader != "myapp.example.com" {
		t.Fatalf("expected X-Forwarded-Host myapp.example.com, got %s", receivedHeader)
	}
}

func TestNew_ModifyRequest(t *testing.T) {
	var receivedToken string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedToken = r.Header.Get("X-Internal-Token")
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	opts := proxy.DefaultOptions(backend.URL)
	opts.ModifyRequest = func(r *http.Request) {
		r.Header.Set("X-Internal-Token", "secret")
	}
	mw := proxy.New(opts)
	handler := mw(http.NotFoundHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if receivedToken != "secret" {
		t.Fatalf("expected X-Internal-Token secret, got %s", receivedToken)
	}
}

func TestNew_PanicsOnEmptyTarget(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on empty target")
		}
	}()
	proxy.New(proxy.Options{})
}
