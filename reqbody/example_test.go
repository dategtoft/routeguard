package reqbody_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/patrickward/routeguard/reqbody"
)

func ExampleNew() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := reqbody.FromContext(r.Context())
		var payload map[string]string
		if err := json.Unmarshal(body, &payload); err == nil {
			fmt.Fprintf(w, "name=%s", payload["name"])
		}
	})

	mw := reqbody.New(reqbody.DefaultOptions())
	ts := httptest.NewServer(mw(handler))
	defer ts.Close()

	resp, _ := http.Post(ts.URL, "application/json",
		strings.NewReader(`{"name":"routeguard"}`))
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	// Output:
	// 200
}

func ExampleNew_customLimit() {
	opts := reqbody.Options{
		MaxBytes: 512,
		OnRead: func(r *http.Request, body []byte) {
			_ = body // log or audit the raw payload
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	mw := reqbody.New(opts)
	ts := httptest.NewServer(mw(handler))
	defer ts.Close()

	resp, _ := http.Post(ts.URL, "text/plain", strings.NewReader("small"))
	defer resp.Body.Close()
	fmt.Println(resp.StatusCode)
	// Output:
	// 204
}
