package jwt

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims used by routeguard.
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// Validator holds configuration for JWT validation.
type Validator struct {
	secret []byte
}

// New creates a new JWT Validator with the provided secret key.
func New(secret string) *Validator {
	return &Validator{secret: []byte(secret)}
}

// Parse extracts and validates a JWT from the Authorization header.
// Returns the parsed Claims or an error.
func (v *Validator) Parse(r *http.Request) (*Claims, error) {
	token := extractToken(r)
	if token == "" {
		return nil, errors.New("missing authorization token")
	}

	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return v.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

// Middleware returns an HTTP middleware that validates JWT tokens.
// Requests with invalid or missing tokens receive a 401 Unauthorized response.
func (v *Validator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := v.Parse(r)
		if err != nil {
			http.Error(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// GenerateToken creates a signed JWT string for the given userID with the specified TTL.
func (v *Validator) GenerateToken(userID string, ttl time.Duration) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(v.secret)
}

// extractToken retrieves the Bearer token from the Authorization header.
func extractToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	parts := strings.SplitN(header, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
		return parts[1]
	}
	return ""
}
