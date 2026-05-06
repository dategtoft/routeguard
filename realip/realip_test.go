package realip

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newCaptureHandler(addr *string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*addr = r.RemoteAddr
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if len(opts.TrustedProxies) == 0 {
		t.Fatal("expected default trusted proxies")
	}
	if len(opts.Headers) == 0 {
		t.Fatal("expected default headers")
	}
}

func TestNew_UntrustedProxy_KeepsRemoteAddr(t *testing.T) {
	var got string
	h := New(DefaultOptions())(newCaptureHandler(&got))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:9000" // not in trusted ranges
	req.Header.Set("X-Forwarded-For", "5.6.7.8")

	h.ServeHTTP(httptest.NewRecorder(), req)

	if got != "1.2.3.4:9000" {
		t.Fatalf("expected RemoteAddr unchanged, got %s", got)
	}
}

func TestNew_TrustedProxy_XForwardedFor(t *testing.T) {
	var got string
	h := New(DefaultOptions())(newCaptureHandler(&got))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:9000"
	req.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.1")

	h.ServeHTTP(httptest.NewRecorder(), req)

	if got != "203.0.113.5:9000" {
		t.Fatalf("expected real IP, got %s", got)
	}
}

func TestNew_TrustedProxy_XRealIP(t *testing.T) {
	var got string
	h := New(DefaultOptions())(newCaptureHandler(&got))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.2:4321"
	req.Header.Set("X-Real-IP", "198.51.100.7")

	h.ServeHTTP(httptest.NewRecorder(), req)

	if got != "198.51.100.7:4321" {
		t.Fatalf("expected real IP, got %s", got)
	}
}

func TestNew_TrustedProxy_NoHeader_KeepsRemoteAddr(t *testing.T) {
	var got string
	h := New(DefaultOptions())(newCaptureHandler(&got))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:8080"
	// no forwarding headers

	h.ServeHTTP(httptest.NewRecorder(), req)

	if got != "192.168.1.1:8080" {
		t.Fatalf("expected RemoteAddr unchanged, got %s", got)
	}
}

func TestNew_CustomHeader(t *testing.T) {
	var got string
	opts := Options{
		TrustedProxies: []string{"127.0.0.0/8"},
		Headers:        []string{"CF-Connecting-IP"},
	}
	h := New(opts)(newCaptureHandler(&got))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("CF-Connecting-IP", "203.0.113.99")

	h.ServeHTTP(httptest.NewRecorder(), req)

	if got != "203.0.113.99:1234" {
		t.Fatalf("expected CF IP, got %s", got)
	}
}
