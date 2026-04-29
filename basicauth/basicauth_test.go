package basicauth

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func basicAuthHeader(user, pass string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+pass))
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.Realm != "Restricted" {
		t.Errorf("expected realm 'Restricted', got %q", opts.Realm)
	}
	if opts.Credentials == nil {
		t.Error("expected non-nil credentials map")
	}
}

func TestNew_ValidCredentials(t *testing.T) {
	opts := DefaultOptions()
	opts.Credentials["admin"] = "secret"
	mw := New(opts)(newTestHandler())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", basicAuthHeader("admin", "secret"))
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestNew_WrongPassword(t *testing.T) {
	opts := DefaultOptions()
	opts.Credentials["admin"] = "secret"
	mw := New(opts)(newTestHandler())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", basicAuthHeader("admin", "wrong"))
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestNew_MissingHeader(t *testing.T) {
	opts := DefaultOptions()
	opts.Credentials["admin"] = "secret"
	mw := New(opts)(newTestHandler())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
	if rr.Header().Get("WWW-Authenticate") == "" {
		t.Error("expected WWW-Authenticate header to be set")
	}
}

func TestNew_UnknownUser(t *testing.T) {
	opts := DefaultOptions()
	opts.Credentials["admin"] = "secret"
	mw := New(opts)(newTestHandler())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", basicAuthHeader("ghost", "secret"))
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestNew_CustomUnauthorizedHandler(t *testing.T) {
	opts := DefaultOptions()
	opts.Credentials["admin"] = "secret"
	opts.UnauthorizedHandler = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}
	mw := New(opts)(newTestHandler())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}
