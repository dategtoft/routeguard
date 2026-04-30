package csrf_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/routeguard/csrf"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := csrf.DefaultOptions()
	if opts.TokenHeader == "" {
		t.Error("expected non-empty TokenHeader")
	}
	if opts.CookieName == "" {
		t.Error("expected non-empty CookieName")
	}
	if opts.TokenLength == 0 {
		t.Error("expected non-zero TokenLength")
	}
	if len(opts.SafeMethods) == 0 {
		t.Error("expected at least one safe method")
	}
}

func TestNew_SafeMethod_IssuesCookie(t *testing.T) {
	handler := csrf.New(csrf.DefaultOptions())(newTestHandler())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	cookies := rec.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == "csrf_token" {
			found = true
			if c.Value == "" {
				t.Error("expected non-empty cookie value")
			}
		}
	}
	if !found {
		t.Error("expected csrf_token cookie to be set")
	}
}

func TestNew_SafeMethod_ExistingCookieNotReplaced(t *testing.T) {
	handler := csrf.New(csrf.DefaultOptions())(newTestHandler())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "existing-token"})
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	// No new cookie should be set
	for _, c := range rec.Result().Cookies() {
		if c.Name == "csrf_token" {
			t.Error("should not replace existing csrf_token cookie")
		}
	}
}

func TestNew_UnsafeMethod_MissingCookie_Returns403(t *testing.T) {
	handler := csrf.New(csrf.DefaultOptions())(newTestHandler())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestNew_UnsafeMethod_TokenMismatch_Returns403(t *testing.T) {
	handler := csrf.New(csrf.DefaultOptions())(newTestHandler())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "valid-token"})
	req.Header.Set("X-CSRF-Token", "wrong-token")
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestNew_UnsafeMethod_ValidToken_Passes(t *testing.T) {
	handler := csrf.New(csrf.DefaultOptions())(newTestHandler())

	const token = "abc123"
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: token})
	req.Header.Set("X-CSRF-Token", token)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestNew_CustomOptions(t *testing.T) {
	opts := csrf.Options{
		TokenHeader: "X-My-CSRF",
		CookieName:  "my_csrf",
		TokenLength: 16,
		SafeMethods: []string{http.MethodGet},
	}
	handler := csrf.New(opts)(newTestHandler())

	const token = "custom-token-value"
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	req.AddCookie(&http.Cookie{Name: "my_csrf", Value: token})
	req.Header.Set("X-My-CSRF", token)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
