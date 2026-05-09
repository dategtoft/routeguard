# geo

Middleware that filters HTTP requests by geographic location using a caller-supplied IP-to-country lookup function.

## Usage

```go
import "github.com/yourusername/routeguard/geo"

// Provide your own GeoIP lookup (e.g. MaxMind GeoLite2).
lookup := func(ip string) string {
    // return ISO 3166-1 alpha-2 country code
    return myGeoIPDB.Lookup(ip)
}

opts := geo.DefaultOptions()
opts.Lookup = lookup
opts.Allowlist = []string{"US", "CA", "GB"}

mw := geo.New(opts)
http.Handle("/", mw(myHandler))
```

## Options

| Field | Type | Default | Description |
|---|---|---|---|
| `Lookup` | `LookupFunc` | `nil` | Resolves client IP → country code. No-op when nil. |
| `Allowlist` | `[]string` | `nil` | If non-empty, only these country codes are allowed. |
| `Blocklist` | `[]string` | `nil` | If non-empty, these country codes are denied. |
| `DeniedCode` | `int` | `403` | HTTP status returned on denial. |
| `DeniedBody` | `string` | `"Forbidden"` | Response body on denial. |
| `CountryHeader` | `string` | `""` | When set, injects the resolved country code into this request header. |

> **Note:** `Allowlist` takes precedence over `Blocklist` when both are set.

## Country code resolution order

The client IP is resolved using, in order:
1. `X-Real-IP` header
2. First entry of `X-Forwarded-For` header
3. `RemoteAddr`
