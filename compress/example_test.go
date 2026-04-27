package compress_test

import (
	"fmt"
	"net/http"

	"github.com/yourusername/routeguard/compress"
)

// ExampleNew demonstrates how to wrap an HTTP handler with
// gzip compression middleware using default options.
func ExampleNew() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, compressed world!")
	})

	// Use default options: gzip.DefaultCompression, MinLength 1024.
	opts := compress.DefaultOptions()
	compressed := compress.New(opts)(handler)

	http.Handle("/", compressed)
}

// ExampleNew_customLevel demonstrates configuring a custom compression level.
func ExampleNew_customLevel() {
	import_gzip := 1 // gzip.BestSpeed
	_ = import_gzip

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Fast compression!")
	})

	opts := compress.Options{
		Level:     1, // gzip.BestSpeed
		MinLength: 512,
	}
	compressed := compress.New(opts)(handler)

	http.Handle("/fast", compressed)
}
