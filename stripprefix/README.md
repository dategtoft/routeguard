# stripprefix

The `stripprefix` package provides HTTP middleware that removes a configured
URL path prefix before the request reaches the next handler. This is useful
when mounting sub-applications or versioned API groups behind a shared router.

## Usage

```go
import "github.com/yourusername/routeguard/stripprefix"

mw := stripprefix.New(stripprefix.Options{
    Prefix: "/api/v1",
})

http.ListenAndServe(":8080", mw(myHandler))
```

A `GET /api/v1/users` request will reach `myHandler` with the path `/users`.

## Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Prefix` | `string` | `""` | URL path prefix to strip. Leading/trailing slashes are normalised automatically. |
| `RedirectOnMismatch` | `bool` | `false` | Return `404 Not Found` when the request path does not start with `Prefix`. When `false`, the request is passed through unchanged. |

## Chaining with other middleware

```go
chain := middleware.Chain(
    stripprefix.New(stripprefix.Options{Prefix: "/api"}),
    logger.New(logger.DefaultOptions()),
)

http.ListenAndServe(":8080", chain(myHandler))
```
