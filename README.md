# routeguard

Lightweight middleware library for Go HTTP routers with rate limiting and JWT validation built-in.

---

## Installation

```bash
go get github.com/yourusername/routeguard
```

---

## Usage

```go
package main

import (
    "net/http"
    "github.com/yourusername/routeguard"
)

func main() {
    rg := routeguard.New(routeguard.Config{
        JWTSecret:     "your-secret-key",
        RateLimit:     100, // requests per minute
        RateLimitBy:   routeguard.ByIP,
    })

    mux := http.NewServeMux()

    mux.Handle("/api/protected", rg.Chain(
        rg.RateLimit(),
        rg.ValidateJWT(),
    )(http.HandlerFunc(protectedHandler)))

    http.ListenAndServe(":8080", mux)
}

func protectedHandler(w http.ResponseWriter, r *http.Request) {
    claims, _ := routeguard.ClaimsFromContext(r.Context())
    w.Write([]byte("Hello, " + claims.Subject))
}
```

---

## Features

- 🔒 JWT validation with configurable secret and claims parsing
- ⚡ Token bucket rate limiting per IP or API key
- 🔗 Chainable middleware compatible with `net/http` and most routers
- 🪶 Zero heavy dependencies

---

## License

[MIT](LICENSE)