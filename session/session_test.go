package session_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrickward/routeguard/session"
)

var testSecret = []byte("super-secret-key")

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := session.DefaultOptions(testSecret)
	if opts.CookieName != "sid" {
		t.Errorf("expected cookie name 'sid', got %q", opts.CookieName)
	}
	if opts.TTL == 0 {
		t.Error("expected non-zero TTL")
	}
	if !opts.HTTPOnly {
		t.Error("expected HTTPOnly to be true")
	}
}

func TestNew_IssuesCookieOnFirstVisit(t *testing.T) {
	mw := session.New(session.DefaultOptions(testSecret))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	mw(newTestHandler(t)).ServeHTTP(rec, req)

	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected a Set-Cookie header")
	}
	if cookies[0].Name != "sid" {
		t.Errorf("expected cookie name 'sid', got %q", cookies[0].Name)
	}
	if cookies[0].Value == "" {
		t.Error("expected non-empty cookie value")
	}
}

func TestNew_ReusesCookieOnSubsequentVisit(t *testing.T) {
	opts := session.DefaultOptions(testSecret)
	mw := session.New(opts)

	// First request — capture cookie.
	rec1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(newTestHandler(t)).ServeHTTP(rec1, req1)
	cookie := rec1.Result().Cookies()[0]

	// Second request — send the cookie back.
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.AddCookie(cookie)

	var capturedID string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = session.FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})
	mw(handler).ServeHTTP(rec2, req2)

	if capturedID == "" {
		t.Fatal("expected session ID in context")
	}
	if len(rec2.Result().Cookies()) != 0 {
		t.Error("expected no new Set-Cookie on repeat visit")
	}
}

func TestNew_IDStoredInContext(t *testing.T) {
	mw := session.New(session.DefaultOptions(testSecret))
	var gotID string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = session.FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(handler).ServeHTTP(rec, req)

	if gotID == "" {
		t.Error("expected session ID to be stored in context")
	}
}

func TestNew_TamperedCookieIssuesNew(t *testing.T) {
	opts := session.DefaultOptions(testSecret)
	mw := session.New(opts)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "sid", Value: "tampered.invalidsig"})

	mw(newTestHandler(t)).ServeHTTP(rec, req)

	if len(rec.Result().Cookies()) == 0 {
		t.Error("expected a new cookie to be issued for tampered value")
	}
}

func TestNew_CustomCookieName(t *testing.T) {
	opts := session.DefaultOptions(testSecret)
	opts.CookieName = "my_session"
	mw := session.New(opts)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(newTestHandler(t)).ServeHTTP(rec, req)

	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected a Set-Cookie header")
	}
	if cookies[0].Name != "my_session" {
		t.Errorf("expected cookie name 'my_session', got %q", cookies[0].Name)
	}
}
