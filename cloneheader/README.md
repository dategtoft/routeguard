# cloneheader

Middleware that copies values from one request header into another before the request reaches your handler.

## Use-case

Reverse proxies (nginx, Envoy, AWS ALB …) often inject identity or routing information under non-standard header names. `cloneheader` lets you map those names to the canonical names your application expects — without changing any proxy configuration.

## Usage

```go
import "github.com/jaxron/routeguard/cloneheader"

handler := cloneheader.New(
    cloneheader.Options{
        Rules: []cloneheader.Rule{
            {Source: "X-Forwarded-User", Destination: "X-User"},
            {Source: "X-Amzn-Oidc-Identity", Destination: "X-User-ID"},
        },
        // Overwrite: true, // replace destination if already present
    },
    myHandler,
)
```

## Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Rules` | `[]Rule` | `nil` | Ordered list of source→destination header pairs |
| `Overwrite` | `bool` | `false` | Replace destination header when it already has a value |

## Notes

- The original `*http.Request` is **never mutated**; a shallow clone is passed downstream.
- Rules are applied in order; later rules can overwrite destinations set by earlier ones when `Overwrite` is `true`.
- If a source header is absent the rule is silently skipped.
