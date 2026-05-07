package session_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrickward/routeguard/session"
)

func BenchmarkSession_NewCookie(b *testing.B) {
	opts := session.DefaultOptions(testSecret)
	mw := session.New(opts)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkSession_ExistingCookie(b *testing.B) {
	opts := session.DefaultOptions(testSecret)
	mw := session.New(opts)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Obtain a valid cookie first.
	rec0 := httptest.NewRecorder()
	req0 := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec0, req0)
	cookie := rec0.Result().Cookies()[0]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(cookie)
		handler.ServeHTTP(rec, req)
	}
}
