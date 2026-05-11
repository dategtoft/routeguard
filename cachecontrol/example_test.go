package cachecontrol_test

import (
	"net/http"

	"github.com/joeydotdev/routeguard/cachecontrol"
)

// ExampleNew demonstrates applying default Cache-Control headers.
func ExampleNew() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})

	// Wrap with default options: public, max-age=60
	middleware := cachecontrol.New(cachecontrol.DefaultOptions())
	http.Handle("/", middleware(handler))
}

// ExampleNew_immutable shows how to cache static assets indefinitely.
func ExampleNew_immutable() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("static asset"))
	})

	// One year + immutable: ideal for content-addressed assets.
	middleware := cachecontrol.New(cachecontrol.Options{
		MaxAge:    31536000,
		Immutable: true,
	})
	http.Handle("/static/", middleware(handler))
}

// ExampleNew_private shows how to prevent shared caching for user-specific responses.
func ExampleNew_private() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("user data"))
	})

	middleware := cachecontrol.New(cachecontrol.Options{
		Private: true,
		NoCache: true,
	})
	http.Handle("/profile", middleware(handler))
}
