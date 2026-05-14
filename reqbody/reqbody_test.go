package reqbody_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/patrickward/routeguard/reqbody"
)

func newTestHandler(t *testing.T, wantBody string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := reqbody.FromContext(r.Context())
		if string(got) != wantBody {
			t.Errorf("FromContext: got %q, want %q", got, wantBody)
		}
		// Ensure body is still readable.
		raw, _ := io.ReadAll(r.Body)
		if string(raw) != wantBody {
			t.Errorf("r.Body: got %q, want %q", raw, wantBody)
		}
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := reqbody.DefaultOptions()
	if opts.MaxBytes != 1<<20 {
		t.Errorf("MaxBytes = %d, want %d", opts.MaxBytes, 1<<20)
	}
}

func TestNew_BuffersBody(t *testing.T) {
	body := `{"hello":"world"}`
	mw := reqbody.New(reqbody.DefaultOptions())
	h := mw(newTestHandler(t, body))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestNew_NilBody_PassesThrough(t *testing.T) {
	mw := reqbody.New(reqbody.DefaultOptions())
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Body = nil
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestNew_OnRead_Callback(t *testing.T) {
	var captured []byte
	opts := reqbody.DefaultOptions()
	opts.OnRead = func(_ *http.Request, b []byte) { captured = b }

	mw := reqbody.New(opts)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) }))

	body := "callback-test"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	h.ServeHTTP(httptest.NewRecorder(), req)

	if string(captured) != body {
		t.Errorf("OnRead captured %q, want %q", captured, body)
	}
}

func TestNew_MaxBytes_Truncates(t *testing.T) {
	opts := reqbody.Options{MaxBytes: 5}
	mw := reqbody.New(opts)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := reqbody.FromContext(r.Context())
		if len(got) > 5 {
			t.Errorf("body length %d exceeds MaxBytes 5", len(got))
		}
		w.WriteHeader(200)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("hello world"))
	h.ServeHTTP(httptest.NewRecorder(), req)
}
