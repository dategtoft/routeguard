package mimetype_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/patrickward/routeguard/mimetype"
)

func ExampleNew() {
	// Allow only JSON responses from downstream handlers.
	mw := mimetype.New(mimetype.Options{
		Allowed: []string{"application/json"},
	})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"ok":true}`)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	handler.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	// Output: 200
}

func ExampleNew_wildcardType() {
	// Allow any image/* response.
	mw := mimetype.New(mimetype.Options{
		Allowed: []string{"image/*"},
	})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/photo.jpg", nil)
	handler.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	// Output: 200
}
