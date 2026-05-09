package geo_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/routeguard/geo"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func staticLookup(country string) geo.LookupFunc {
	return func(_ string) string { return country }
}

func TestDefaultOptions(t *testing.T) {
	opts := geo.DefaultOptions()
	if opts.DeniedCode != http.StatusForbidden {
		t.Errorf("expected 403, got %d", opts.DeniedCode)
	}
	if opts.DeniedBody != "Forbidden" {
		t.Errorf("unexpected body: %s", opts.DeniedBody)
	}
}

func TestNew_NilLookup_PassesThrough(t *testing.T) {
	mw := geo.New(geo.DefaultOptions())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(newTestHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestNew_Allowlist_AllowedCountry(t *testing.T) {
	opts := geo.DefaultOptions()
	opts.Lookup = staticLookup("US")
	opts.Allowlist = []string{"US", "CA"}
	mw := geo.New(opts)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(newTestHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestNew_Allowlist_BlockedCountry(t *testing.T) {
	opts := geo.DefaultOptions()
	opts.Lookup = staticLookup("CN")
	opts.Allowlist = []string{"US", "CA"}
	mw := geo.New(opts)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(newTestHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestNew_Blocklist_BlockedCountry(t *testing.T) {
	opts := geo.DefaultOptions()
	opts.Lookup = staticLookup("RU")
	opts.Blocklist = []string{"RU", "KP"}
	mw := geo.New(opts)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(newTestHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestNew_Blocklist_AllowedCountry(t *testing.T) {
	opts := geo.DefaultOptions()
	opts.Lookup = staticLookup("DE")
	opts.Blocklist = []string{"RU", "KP"}
	mw := geo.New(opts)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(newTestHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestNew_CountryHeader_Set(t *testing.T) {
	var capturedCountry string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCountry = r.Header.Get("X-Country")
		w.WriteHeader(http.StatusOK)
	})

	opts := geo.DefaultOptions()
	opts.Lookup = staticLookup("JP")
	opts.CountryHeader = "X-Country"
	mw := geo.New(opts)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(handler).ServeHTTP(rec, req)

	if capturedCountry != "JP" {
		t.Errorf("expected JP, got %q", capturedCountry)
	}
}

func TestNew_CaseInsensitiveCountryCodes(t *testing.T) {
	opts := geo.DefaultOptions()
	opts.Lookup = staticLookup("us") // lowercase from lookup
	opts.Allowlist = []string{"US"}   // uppercase in list
	mw := geo.New(opts)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(newTestHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
