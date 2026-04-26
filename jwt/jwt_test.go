package jwt

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testSecret = "super-secret-key"

func TestGenerateAndParse_Valid(t *testing.T) {
	v := New(testSecret)
	token, err := v.GenerateToken("user-42", time.Hour)
	if err != nil {
		t.Fatalf("expected no error generating token, got: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	claims, err := v.Parse(req)
	if err != nil {
		t.Fatalf("expected valid token, got error: %v", err)
	}
	if claims.UserID != "user-42" {
		t.Errorf("expected user_id 'user-42', got '%s'", claims.UserID)
	}
}

func TestParse_MissingToken(t *testing.T) {
	v := New(testSecret)
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	_, err := v.Parse(req)
	if err == nil {
		t.Fatal("expected error for missing token")
	}
}

func TestParse_ExpiredToken(t *testing.T) {
	v := New(testSecret)
	token, _ := v.GenerateToken("user-1", -time.Second)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	_, err := v.Parse(req)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestParse_InvalidSecret(t *testing.T) {
	v := New(testSecret)
	token, _ := v.GenerateToken("user-1", time.Hour)

	wrongValidator := New("wrong-secret")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	_, err := wrongValidator.Parse(req)
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestMiddleware_UnauthorizedRequest(t *testing.T) {
	v := New(testSecret)
	handler := v.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestMiddleware_AuthorizedRequest(t *testing.T) {
	v := New(testSecret)
	token, _ := v.GenerateToken("user-99", time.Hour)

	handler := v.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
