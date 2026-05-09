package signature_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joeydtaylor/routeguard/signature"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	o := signature.DefaultOptions()
	if o.Header != "X-Signature" {
		t.Errorf("expected header X-Signature, got %s", o.Header)
	}
	if o.Prefix != "sha256=" {
		t.Errorf("expected prefix sha256=, got %s", o.Prefix)
	}
}

func TestNew_ValidSignature_Passes(t *testing.T) {
	secret := "supersecret"
	mw := signature.New(secret)
	handler := mw(newTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	sig := "sha256=" + signature.Sign(secret, http.MethodGet, "/hello")
	req.Header.Set("X-Signature", sig)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestNew_MissingSignature_Returns401(t *testing.T) {
	mw := signature.New("secret")
	handler := mw(newTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestNew_InvalidSignature_Returns401(t *testing.T) {
	mw := signature.New("secret")
	handler := mw(newTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	req.Header.Set("X-Signature", "sha256=badhex")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestNew_WrongSecret_Returns401(t *testing.T) {
	mw := signature.New("correct-secret")
	handler := mw(newTestHandler())

	req := httptest.NewRequest(http.MethodPost, "/data", nil)
	sig := "sha256=" + signature.Sign("wrong-secret", http.MethodPost, "/data")
	req.Header.Set("X-Signature", sig)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestNew_CustomHeader_NoPrefix(t *testing.T) {
	secret := "mysecret"
	opts := signature.Options{
		Header: "X-Hub-Signature",
		Prefix: "",
	}
	mw := signature.New(secret, opts)
	handler := mw(newTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	sig := signature.Sign(secret, http.MethodGet, "/ping")
	req.Header.Set("X-Hub-Signature", sig)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
