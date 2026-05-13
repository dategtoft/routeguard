package acceptlang_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joeydtaylor/routeguard/acceptlang"
)

func newTestHandler(t *testing.T, wantLang string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := acceptlang.FromContext(r.Context())
		if got != wantLang {
			t.Errorf("FromContext = %q; want %q", got, wantLang)
		}
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := acceptlang.DefaultOptions()
	if opts.Default != "en" {
		t.Errorf("Default = %q; want \"en\"", opts.Default)
	}
	if opts.Header != "Content-Language" {
		t.Errorf("Header = %q; want \"Content-Language\"", opts.Header)
	}
}

func TestNew_NoHeader_UsesDefault(t *testing.T) {
	mw := acceptlang.New(acceptlang.DefaultOptions())
	h := mw(newTestHandler(t, "en"))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d; want 200", rec.Code)
	}
}

func TestNew_MatchedLanguage(t *testing.T) {
	opts := acceptlang.DefaultOptions()
	opts.Supported = []string{"en", "fr", "de"}
	mw := acceptlang.New(opts)
	h := mw(newTestHandler(t, "fr"))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8")
	h.ServeHTTP(rec, req)
}

func TestNew_BaseLanguageFallback(t *testing.T) {
	opts := acceptlang.DefaultOptions()
	opts.Supported = []string{"en", "fr"}
	mw := acceptlang.New(opts)
	h := mw(newTestHandler(t, "fr"))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "fr-CA")
	h.ServeHTTP(rec, req)
}

func TestNew_UnsupportedLanguage_UsesDefault(t *testing.T) {
	opts := acceptlang.DefaultOptions()
	opts.Supported = []string{"en"}
	mw := acceptlang.New(opts)
	h := mw(newTestHandler(t, "en"))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "ja, zh")
	h.ServeHTTP(rec, req)
}

func TestNew_SetsResponseHeader(t *testing.T) {
	opts := acceptlang.DefaultOptions()
	opts.Supported = []string{"en", "de"}
	mw := acceptlang.New(opts)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "de")
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get("Content-Language"); got != "de" {
		t.Errorf("Content-Language = %q; want \"de\"", got)
	}
}

func TestNew_NoSupportedList_AcceptsAny(t *testing.T) {
	opts := acceptlang.DefaultOptions()
	mw := acceptlang.New(opts)
	h := mw(newTestHandler(t, "ja"))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "ja")
	h.ServeHTTP(rec, req)
}
