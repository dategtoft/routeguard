# cache

The `cache` package provides an in-memory HTTP response caching middleware for Go HTTP routers.

## Features

- Caches responses in memory with a configurable TTL.
- Only caches configurable HTTP methods (default: `GET`, `HEAD`).
- Adds `X-Cache: HIT` or `X-Cache: MISS` headers to every response.
- Safe for concurrent use.

## Usage

```go
import "github.com/yourusername/routeguard/cache"

// Use default options (60s TTL, GET and HEAD cached).
handler := cache.New(cache.DefaultOptions())(myHandler)
http.ListenAndServe(":8080", handler)
```

## Custom Options

```go
opts := cache.Options{
    TTL:     30 * time.Second,
    Methods: []string{http.MethodGet},
}
handler := cache.New(opts)(myHandler)
```

## Response Headers

| Header    | Value  | Meaning                          |
|-----------|--------|----------------------------------|
| `X-Cache` | `HIT`  | Response served from cache       |
| `X-Cache` | `MISS` | Response fetched from the handler|

## Notes

- The cache key is the full request URI (path + query string).
- Non-cacheable methods (e.g. `POST`, `PUT`) are always forwarded to the next handler.
- Expired entries are evicted lazily on the next request for the same key.
