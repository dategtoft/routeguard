// Package signature provides HTTP request signature verification middleware.
// It validates HMAC-signed requests using a shared secret and a configurable
// header, rejecting any request whose signature does not match.
package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
)

// Options configures the signature middleware.
type Options struct {
	// Secret is the shared HMAC secret used to verify request signatures.
	Secret string
	// Header is the HTTP header that carries the signature.
	// Defaults to "X-Signature".
	Header string
	// Prefix is an optional prefix stripped from the header value before
	// verification (e.g. "sha256=").
	Prefix string
}

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Header: "X-Signature",
		Prefix: "sha256=",
	}
}

// New returns middleware that verifies HMAC-SHA256 request signatures.
// Requests missing or carrying an invalid signature receive 401 Unauthorized.
func New(secret string, opts ...Options) func(http.Handler) http.Handler {
	o := DefaultOptions()
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Header == "" {
		o.Header = "X-Signature"
	}
	o.Secret = secret

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := r.Header.Get(o.Header)
			if raw == "" {
				http.Error(w, "missing signature", http.StatusUnauthorized)
				return
			}

			sig := raw
			if o.Prefix != "" {
				sig = strings.TrimPrefix(raw, o.Prefix)
			}

			payload := r.Method + r.URL.RequestURI()
			if !verify(o.Secret, payload, sig) {
				http.Error(w, "invalid signature", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Sign returns the HMAC-SHA256 hex signature for the given method and URI
// using secret. The result does not include any prefix.
func Sign(secret, method, uri string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(method + uri))
	return hex.EncodeToString(mac.Sum(nil))
}

func verify(secret, payload, sig string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	expected := mac.Sum(nil)
	actual, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}
	return hmac.Equal(expected, actual)
}
