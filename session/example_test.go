package session_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/patrickward/routeguard/session"
)

// ExampleNew demonstrates attaching the session middleware to a handler.
func ExampleNew() {
	secret := []byte("my-secret-key")
	opts := session.DefaultOptions(secret)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := session.FromContext(r.Context())
		fmt.Fprintf(w, "session: %s", id)
	})

	mw := session.New(opts)
	ts := httptest.NewServer(mw(handler))
	defer ts.Close()

	resp, _ := http.Get(ts.URL)
	defer resp.Body.Close()

	cookies := resp.Cookies()
	if len(cookies) > 0 {
		fmt.Println("cookie issued:", cookies[0].Name)
	}
	// Output:
	// cookie issued: sid
}

// ExampleNew_customOptions demonstrates using a custom cookie name and Secure flag.
func ExampleNew_customOptions() {
	secret := []byte("my-secret-key")
	opts := session.DefaultOptions(secret)
	opts.CookieName = "app_session"
	opts.Secure = true

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := session.New(opts)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(handler).ServeHTTP(rec, req)

	cookies := rec.Result().Cookies()
	if len(cookies) > 0 {
		fmt.Println("cookie name:", cookies[0].Name)
		fmt.Println("secure:", cookies[0].Secure)
	}
	// Output:
	// cookie name: app_session
	// secure: true
}
