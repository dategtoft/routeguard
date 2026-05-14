# slowdown

The `slowdown` middleware introduces an artificial delay before passing requests to the next handler. It is intended for **development and testing** scenarios where you want to simulate network latency or slow upstream services.

## Usage

```go
import "github.com/jcosta33/routeguard/slowdown"

// Delay all requests by 300ms
mw := slowdown.New(slowdown.Options{
    Delay: 300 * time.Millisecond,
})

http.Handle("/", mw(myHandler))
```

## Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Delay` | `time.Duration` | `500ms` | How long to wait before forwarding the request. |
| `OnlyPaths` | `[]string` | `nil` | If set, only requests to these exact paths are delayed. |

## Context cancellation

If the request context is cancelled while the middleware is waiting (e.g. the client disconnects), the middleware responds immediately with `503 Service Unavailable` instead of blocking indefinitely.

## Default options

```go
opts := slowdown.DefaultOptions()
// opts.Delay     == 500ms
// opts.OnlyPaths == nil  (all paths delayed)
```
