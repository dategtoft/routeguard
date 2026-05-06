package paginate_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/routeguard/paginate"
)

func newTestHandler(t *testing.T, check func(p paginate.Params)) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		check(paginate.FromContext(r.Context()))
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := paginate.DefaultOptions()
	if opts.DefaultLimit != 20 {
		t.Errorf("expected DefaultLimit 20, got %d", opts.DefaultLimit)
	}
	if opts.MaxLimit != 100 {
		t.Errorf("expected MaxLimit 100, got %d", opts.MaxLimit)
	}
}

func TestNew_DefaultParams(t *testing.T) {
	mw := paginate.New(paginate.DefaultOptions())
	handler := mw(newTestHandler(t, func(p paginate.Params) {
		if p.Page != 1 {
			t.Errorf("expected page 1, got %d", p.Page)
		}
		if p.Limit != 20 {
			t.Errorf("expected limit 20, got %d", p.Limit)
		}
		if p.Offset != 0 {
			t.Errorf("expected offset 0, got %d", p.Offset)
		}
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestNew_CustomPageAndLimit(t *testing.T) {
	mw := paginate.New(paginate.DefaultOptions())
	handler := mw(newTestHandler(t, func(p paginate.Params) {
		if p.Page != 3 {
			t.Errorf("expected page 3, got %d", p.Page)
		}
		if p.Limit != 10 {
			t.Errorf("expected limit 10, got %d", p.Limit)
		}
		if p.Offset != 20 {
			t.Errorf("expected offset 20, got %d", p.Offset)
		}
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/items?page=3&limit=10", nil)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestNew_LimitCappedAtMax(t *testing.T) {
	opts := paginate.DefaultOptions()
	opts.MaxLimit = 50
	mw := paginate.New(opts)
	handler := mw(newTestHandler(t, func(p paginate.Params) {
		if p.Limit != 50 {
			t.Errorf("expected limit capped at 50, got %d", p.Limit)
		}
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/items?limit=999", nil)
	handler.ServeHTTP(rec, req)
}

func TestNew_InvalidPage_Returns400(t *testing.T) {
	mw := paginate.New(paginate.DefaultOptions())
	handler := mw(newTestHandler(t, func(p paginate.Params) {}))

	for _, bad := range []string{"0", "-1", "abc"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/items?page="+bad, nil)
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("page=%q: expected 400, got %d", bad, rec.Code)
		}
	}
}

func TestNew_InvalidLimit_Returns400(t *testing.T) {
	mw := paginate.New(paginate.DefaultOptions())
	handler := mw(newTestHandler(t, func(p paginate.Params) {}))

	for _, bad := range []string{"0", "-5", "xyz"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/items?limit="+bad, nil)
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("limit=%q: expected 400, got %d", bad, rec.Code)
		}
	}
}
