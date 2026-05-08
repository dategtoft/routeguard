# fingerprint

The `fingerprint` middleware generates a SHA-256 request fingerprint from configurable request attributes and makes it available via a response header and the request context.

## Usage

```go
import "github.com/joeydtaylor/routeguard/fingerprint"

mw := fingerprint.New(fingerprint.DefaultOptions())
http.Handle("/", mw(myHandler))
```

## Options

| Field | Default | Description |
|---|---|---|
| `Header` | `X-Request-Fingerprint` | Response header name for the fingerprint |
| `IncludeIP` | `true` | Include the client IP in the fingerprint |
| `IncludeUserAgent` | `true` | Include the `User-Agent` header |
| `ExtraHeaders` | `nil` | Additional request headers to fold in |

## Reading from context

```go
fp := fingerprint.FromContext(r.Context())
```

## Custom options

```go
opts := fingerprint.DefaultOptions()
opts.ExtraHeaders = []string{"X-Tenant-ID", "Accept-Language"}
opts.Header = "X-FP"

mw := fingerprint.New(opts)
```
