# querylog

Middleware that logs query parameters from incoming HTTP requests.

## Usage

```go
import "github.com/patrickward/routeguard/querylog"

opts := querylog.DefaultOptions()
opts.RedactKeys = []string{"token", "api_key"}

http.Handle("/", querylog.New(opts)(myHandler))
```

## Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Logger` | `*log.Logger` | stderr | Logger to write output to |
| `Prefix` | `string` | `[querylog]` | Log line prefix |
| `SkipPaths` | `[]string` | `nil` | Paths to skip logging |
| `RedactKeys` | `[]string` | `nil` | Query keys whose values are redacted |

## Notes

- Requests with no query parameters produce no log output.
- Redaction is case-insensitive: `Token`, `token`, and `TOKEN` are all redacted.
- Skipped paths are matched exactly against `r.URL.Path`.
