# reqbody

Middleware that buffers the incoming HTTP request body and makes it available
to downstream handlers via the request context.

## Features

- Buffers the request body up to a configurable byte limit
- Restores `r.Body` so downstream handlers can still read it normally
- Stores the raw bytes in the context for inspection without re-reading
- Optional `OnRead` callback for logging or auditing

## Usage

```go
import "github.com/patrickward/routeguard/reqbody"

mw := reqbody.New(reqbody.DefaultOptions())
http.Handle("/", mw(myHandler))
```

### Accessing the body in a handler

```go
func myHandler(w http.ResponseWriter, r *http.Request) {
    raw := reqbody.FromContext(r.Context()) // []byte
    // r.Body is still readable as usual
}
```

### Custom options

```go
opts := reqbody.Options{
    MaxBytes: 64 * 1024, // 64 KiB
    OnRead: func(r *http.Request, body []byte) {
        log.Printf("body: %s", body)
    },
}
mw := reqbody.New(opts)
```

## Defaults

| Option     | Default |
|------------|---------|
| `MaxBytes` | 1 MiB   |
| `OnRead`   | nil     |
