package circuitbreaker_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/leodahal4/routeguard/circuitbreaker"
)

func ExampleNew() {
	// Wrap a handler with default circuit breaker settings.
	// The circuit opens after 5 consecutive 5xx responses.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	mw := circuitbreaker.New(circuitbreaker.DefaultOptions())(handler)

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	fmt.Println(rec.Code)
	// Output: 200
}

func ExampleNew_customOptions() {
	// Trip the circuit after 2 failures and recover after 100ms.
	opts := circuitbreaker.Options{
		Threshold:  2,
		Timeout:    100 * time.Millisecond,
		StatusCode: http.StatusServiceUnavailable,
		Message:    "circuit open, try again later",
	}

	failHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	mw := circuitbreaker.New(opts)(failHandler)

	// Two failures trip the circuit.
	for i := 0; i < 2; i++ {
		mw.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// Third request is rejected by the open circuit.
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	fmt.Println(rec.Code)
	// Output: 503
}
