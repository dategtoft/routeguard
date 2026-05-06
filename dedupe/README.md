# dedupe

The `dedupe` package provides HTTP middleware that coalesces concurrent
duplicate requests into a single upstream call. All waiting requests
receive the same response once the upstream handler returns.

## When to use

Use this middleware in front of expensive or slow handlers (e.g. database
queries, external API calls) to prevent a thundering-herd of identical
requests from overwhelming your backend.

## Usage

```go
import "github.com/patrickward/routeguard/dedupe"

mux := http.NewServeMux()
mux.Handle("/api/data", dedupe.New(dedupe.DefaultOptions())(myHandler))
```

## Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `KeyFunc` | `func(*http.Request) string` | `METHOD:RequestURI` | Derives the deduplication key |

## Custom key function

```go
opts := dedupe.Options{
    KeyFunc: func(r *http.Request) string {
        // Deduplicate by path only, ignoring query string.
        return r.Method + ":" + r.URL.Path
    },
}
mw := dedupe.New(opts)
```

## Behaviour

- Only **concurrent** requests are coalesced. Sequential requests each
  trigger their own upstream call.
- The response body, status code, and headers captured from the first
  call are replayed to all waiting callers.
- Requests with different keys are always forwarded independently.
