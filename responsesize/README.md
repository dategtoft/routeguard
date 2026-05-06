# responsesize

Middleware that caps the size of HTTP response bodies written by downstream handlers.

If a handler writes more bytes than the configured `MaxBytes` limit, the response is
replaced with an HTTP 500 and a plain-text error message. Bytes already flushed before
the limit is hit are discarded in favour of the error response.

## Usage

```go
import "github.com/patrickward/routeguard/responsesize"

// Default: 10 MB limit
mw := responsesize.New(responsesize.DefaultOptions())

// Custom limit
mw := responsesize.New(responsesize.Options{
    MaxBytes:     1 * 1024 * 1024, // 1 MB
    ErrorMessage: "response too large",
})

http.Handle("/", mw(yourHandler))
```

## Options

| Field | Type | Default | Description |
|---|---|---|---|
| `MaxBytes` | `int64` | `10485760` (10 MB) | Maximum response body size in bytes. |
| `ErrorMessage` | `string` | `"response body too large"` | Body returned when the limit is exceeded. |

## Notes

- A `MaxBytes` value of `0` or negative falls back to the 10 MB default.
- The middleware replaces the entire response (status + body) when the limit is
  exceeded, so downstream `WriteHeader` calls are suppressed after a violation.
