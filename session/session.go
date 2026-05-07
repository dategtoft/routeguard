// Package session provides HTTP session management middleware.
// It issues a signed session cookie on first visit and makes the
// session ID available via the request context on subsequent requests.
package session

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"
)

type contextKey struct{}

// Options configures the session middleware.
type Options struct {
	// CookieName is the name of the session cookie. Defaults to "sid".
	CookieName string
	// Secret is used to sign the session ID. Required.
	Secret []byte
	// TTL is the cookie Max-Age. Defaults to 24 hours.
	TTL time.Duration
	// Secure sets the Secure flag on the cookie.
	Secure bool
	// HTTPOnly sets the HttpOnly flag on the cookie. Defaults to true.
	HTTPOnly bool
	// SameSite sets the SameSite attribute. Defaults to SameSiteLaxMode.
	SameSite http.SameSite
}

// DefaultOptions returns an Options with sensible defaults.
func DefaultOptions(secret []byte) Options {
	return Options{
		CookieName: "sid",
		Secret:     secret,
		TTL:        24 * time.Hour,
		HTTPOnly:   true,
		SameSite:   http.SameSiteLaxMode,
	}
}

// FromContext returns the session ID stored in ctx, or an empty string.
func FromContext(ctx context.Context) string {
	v, _ := ctx.Value(contextKey{}).(string)
	return v
}

// New returns a middleware that manages session cookies.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.CookieName == "" {
		opts.CookieName = "sid"
	}
	if opts.TTL == 0 {
		opts.TTL = 24 * time.Hour
	}
	if opts.SameSite == 0 {
		opts.SameSite = http.SameSiteLaxMode
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var sessionID string

			if c, err := r.Cookie(opts.CookieName); err == nil {
				if id := verify(c.Value, opts.Secret); id != "" {
					sessionID = id
				}
			}

			if sessionID == "" {
				sessionID = newID()
				http.SetCookie(w, &http.Cookie{
					Name:     opts.CookieName,
					Value:    sign(sessionID, opts.Secret),
					MaxAge:   int(opts.TTL.Seconds()),
					Secure:   opts.Secure,
					HttpOnly: opts.HTTPOnly,
					SameSite: opts.SameSite,
				})
			}

			ctx := context.WithValue(r.Context(), contextKey{}, sessionID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func sign(id string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(id))
	return id + "." + hex.EncodeToString(mac.Sum(nil))
}

func verify(signed string, secret []byte) string {
	for i := len(signed) - 1; i >= 0; i-- {
		if signed[i] == '.' {
			id := signed[:i]
			expected := sign(id, secret)
			if hmac.Equal([]byte(signed), []byte(expected)) {
				return id
			}
			return ""
		}
	}
	return ""
}
