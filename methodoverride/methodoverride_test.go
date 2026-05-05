package methodoverride_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/yourusername/routeguard/methodoverride"
)

func newCaptureHandler() (http.Handler, *string) {
	var method string
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		w.WriteHeader(http.StatusOK)
	})
	return h, &method
}

func TestDefaultOptions(t *testing.T) {
	opts := methodoverride.DefaultOptions()
	if opts.Header != "X-HTTP-Method-Override" {
		t.Errorf("expected default header X-HTTP-Method-Override, got %s", opts.Header)
	}
	if opts.FormField != "_method" {
		t.Errorf("expected default form field _method, got %s", opts.FormField)
	}
	if len(opts.Allowed) == 0 {
		t.Error("expected non-empty allowed methods")
	}
}

func TestNew_NonPOST_NotOverridden(t *testing.T) {
	h, method := newCaptureHandler()
	mw := methodoverride.New(methodoverride.DefaultOptions())(h)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-HTTP-Method-Override", "DELETE")
	mw.ServeHTTP(httptest.NewRecorder(), req)

	if *method != http.MethodGet {
		t.Errorf("expected GET, got %s", *method)
	}
}

func TestNew_OverrideViaHeader(t *testing.T) {
	h, method := newCaptureHandler()
	mw := methodoverride.New(methodoverride.DefaultOptions())(h)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-HTTP-Method-Override", "DELETE")
	mw.ServeHTTP(httptest.NewRecorder(), req)

	if *method != http.MethodDelete {
		t.Errorf("expected DELETE, got %s", *method)
	}
}

func TestNew_OverrideViaFormField(t *testing.T) {
	h, method := newCaptureHandler()
	mw := methodoverride.New(methodoverride.DefaultOptions())(h)

	form := url.Values{"_method": {"PATCH"}}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mw.ServeHTTP(httptest.NewRecorder(), req)

	if *method != http.MethodPatch {
		t.Errorf("expected PATCH, got %s", *method)
	}
}

func TestNew_DisallowedMethod_NotOverridden(t *testing.T) {
	h, method := newCaptureHandler()
	mw := methodoverride.New(methodoverride.DefaultOptions())(h)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-HTTP-Method-Override", "CONNECT")
	mw.ServeHTTP(httptest.NewRecorder(), req)

	if *method != http.MethodPost {
		t.Errorf("expected POST (not overridden), got %s", *method)
	}
}

func TestNew_HeaderTakesPrecedenceOverForm(t *testing.T) {
	h, method := newCaptureHandler()
	mw := methodoverride.New(methodoverride.DefaultOptions())(h)

	form := url.Values{"_method": {"PATCH"}}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-HTTP-Method-Override", "PUT")
	mw.ServeHTTP(httptest.NewRecorder(), req)

	if *method != http.MethodPut {
		t.Errorf("expected PUT (header takes precedence), got %s", *method)
	}
}
