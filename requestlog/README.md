# requestlog

Structured single-line HTTP request logging middleware for Go.

## Usage

```go
import "github.com/yourusername/routeguard/requestlog"

mw := requestlog.New(requestlog.DefaultOptions())
http.Handle("/", mw(myHandler))
```

Each request produces a line like:

```
time=2024-01-15T10:30:00Z method=GET path=/api/users status=200 bytes=342 duration=1.23ms
```

## Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Writer` | `io.Writer` | `os.Stdout` | Log destination |
| `TimeFormat` | `string` | `time.RFC3339` | Timestamp format |
| `SkipPaths` | `[]string` | `nil` | Paths to suppress |

## Skip paths

```go
opts := requestlog.DefaultOptions()
opts.SkipPaths = []string{"/healthz", "/readyz"}
mw := requestlog.New(opts)
```
