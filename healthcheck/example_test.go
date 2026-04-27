package healthcheck_test

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/yourusername/routeguard/healthcheck"
)

// ExampleNew demonstrates a basic health check endpoint with no dependency checks.
func ExampleNew() {
	h := healthcheck.New(healthcheck.DefaultOptions())
	http.Handle("/healthz", h)
	// GET /healthz -> 200 {"status":"ok", ...}
	fmt.Println("health endpoint registered")
	// Output: health endpoint registered
}

// ExampleNew_withChecks shows registering named dependency checks.
func ExampleNew_withChecks() {
	opts := healthcheck.DefaultOptions()

	// Simulate a database ping.
	opts.Checks["database"] = func() error {
		// Replace with real DB ping logic.
		return nil
	}

	// Simulate a cache check that is currently failing.
	opts.Checks["cache"] = func() error {
		return errors.New("timeout")
	}

	h := healthcheck.New(opts)
	http.Handle("/healthz", h)
	// GET /healthz -> 503 {"status":"degraded", "checks":{"database":"ok","cache":"timeout"}}
	fmt.Println("health endpoint with checks registered")
	// Output: health endpoint with checks registered
}
