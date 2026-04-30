// Package csrf provides Cross-Site Request Forgery protection middleware.
package csrf

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const (
	defaultTokenHeader = "X-CSRF-Token"
	defaultCookieName  = "csrf_token"
	defaultTokenLength = 32
)

// Options holds configuration for the CSRF middleware.
type Options struct {
	// TokenHeader is the request header checked for the CSRF token.
	TokenHeader string
	// CookieName is the name of the cookie that stores the CSRF token.
	CookieName string
	// TokenLength is the byte length of generated tokens.
	TokenLength int
	// SafeMethods lists HTTP methods exempt from CSRF checks.
	SafeMethods []string
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		TokenHeader: defaultTokenHeader,
		CookieName:  defaultCookieName,
		TokenLength: defaultTokenLength,
		SafeMethods: []string{http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace},
	}
}

// New returns a CSRF protection middleware using the provided options.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.TokenHeader == "" {
		opts.TokenHeader = defaultTokenHeader
	}
	if opts.CookieName == "" {
		opts.CookieName = defaultCookieName
	}
	if opts.TokenLength == 0 {
		opts.TokenLength = defaultTokenLength
	}
	if len(opts.SafeMethods) == 0 {
		opts.SafeMethods = DefaultOptions().SafeMethods
	}

	safe := make(map[string]struct{}, len(opts.SafeMethods))
	for _, m := range opts.SafeMethods {
		safe[m] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Safe methods: issue token if missing, then pass through.
			if _, ok := safe[r.Method]; ok {
				issueTokenIfMissing(w, r, opts)
				next.ServeHTTP(w, r)
				return
			}

			// Unsafe methods: validate token.
			cookie, err := r.Cookie(opts.CookieName)
			if err != nil || cookie.Value == "" {
				http.Error(w, "CSRF cookie missing", http.StatusForbidden)
				return
			}

			headerToken := r.Header.Get(opts.TokenHeader)
			if !secureCompare(headerToken, cookie.Value) {
				http.Error(w, "CSRF token mismatch", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func issueTokenIfMissing(w http.ResponseWriter, r *http.Request, opts Options) {
	if c, err := r.Cookie(opts.CookieName); err == nil && c.Value != "" {
		return
	}
	token := generateToken(opts.TokenLength)
	http.SetCookie(w, &http.Cookie{
		Name:     opts.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: false, // must be readable by JS to set the header
		SameSite: http.SameSiteStrictMode,
	})
}

func generateToken(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func secureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := 0; i < len(a); i++ {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}
