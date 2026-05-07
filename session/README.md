# session

The `session` package provides lightweight HTTP session management middleware for `routeguard`.

It issues a **signed session cookie** on the first visit and makes the session ID available via the request context on all subsequent requests. No server-side storage is required — the session ID is embedded in the cookie and verified with an HMAC-SHA256 signature.

## Usage

```go
import "github.com/patrickward/routeguard/session"

secret := []byte("replace-with-a-strong-random-key")
opts := session.DefaultOptions(secret)

http.Handle("/", session.New(opts)(myHandler))
```

## Reading the Session ID

Use `session.FromContext` inside any downstream handler:

```go
func myHandler(w http.ResponseWriter, r *http.Request) {
    id := session.FromContext(r.Context())
    fmt.Fprintln(w, "Your session:", id)
}
```

## Options

| Field        | Default            | Description                              |
|--------------|--------------------|------------------------------------------|
| `CookieName` | `"sid"`            | Name of the session cookie               |
| `Secret`     | *(required)*       | HMAC signing key                         |
| `TTL`        | `24h`              | Cookie `Max-Age`                         |
| `Secure`     | `false`            | Set the `Secure` cookie flag             |
| `HTTPOnly`   | `true`             | Set the `HttpOnly` cookie flag           |
| `SameSite`   | `SameSiteLaxMode`  | `SameSite` attribute                     |

## Security Notes

- Use a cryptographically random secret of at least 32 bytes.
- Enable `Secure: true` in production to prevent cookie transmission over plain HTTP.
- The middleware does **not** store session data — pair it with a store (e.g. Redis) if you need server-side session values.
