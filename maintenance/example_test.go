package maintenance_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/yourusername/routeguard/maintenance"
)

func ExampleNew() {
	// Create a maintenance middleware that is initially inactive.
	m := maintenance.New(false, maintenance.DefaultOptions())

	app := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := m.Handler(app)

	// Normal operation — request passes through.
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	fmt.Println(rec.Code)

	// Enable maintenance mode.
	m.Enable()

	// Request is now rejected with 503.
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	fmt.Println(rec.Code)

	// Output:
	// 200
	// 503
}

func ExampleNew_customOptions() {
	opts := maintenance.Options{
		Message:      "We'll be back shortly!",
		RetryAfter:   300,
		JSONResponse: false,
	}

	m := maintenance.New(true, opts)

	rec := httptest.NewRecorder()
	m.Handler(http.NotFoundHandler()).ServeHTTP(
		rec,
		httptest.NewRequest(http.MethodGet, "/", nil),
	)

	fmt.Println(rec.Code)
	fmt.Println(rec.Header().Get("Retry-After"))
	fmt.Println(rec.Body.String())

	// Output:
	// 503
	// 300
	// We'll be back shortly!
}
