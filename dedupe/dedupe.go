// Package dedupe provides middleware that suppresses duplicate concurrent
// requests to the same key (URL path + method), coalescing them into a
// single upstream call and fanning the response out to all waiters.
package dedupe

import (
	"net/http"
	"sync"
)

// Options configures the deduplication middleware.
type Options struct {
	// KeyFunc derives the deduplication key from a request.
	// Defaults to "METHOD:URL.RequestURI".
	KeyFunc func(r *http.Request) string
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		KeyFunc: func(r *http.Request) string {
			return r.Method + ":" + r.URL.RequestURI()
		},
	}
}

// flight represents an in-progress upstream request.
type flight struct {
	wg     sync.WaitGroup
	code   int
	header http.Header
	body   []byte
}

// group manages in-flight requests.
type group struct {
	mu      sync.Mutex
	flights map[string]*flight
}

func (g *group) do(key string, fn func() *flight) *flight {
	g.mu.Lock()
	if f, ok := g.flights[key]; ok {
		g.mu.Unlock()
		f.wg.Wait()
		return f
	}
	f := &flight{}
	f.wg.Add(1)
	g.flights[key] = f
	g.mu.Unlock()

	result := fn()
	f.code = result.code
	f.header = result.header
	f.body = result.body
	f.wg.Done()

	g.mu.Lock()
	delete(g.flights, key)
	g.mu.Unlock()
	return f
}

// New returns middleware that deduplicates concurrent identical requests.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.KeyFunc == nil {
		opts.KeyFunc = DefaultOptions().KeyFunc
	}
	g := &group{flights: make(map[string]*flight)}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := opts.KeyFunc(r)
			f := g.do(key, func() *flight {
				rec := &recorder{code: http.StatusOK, header: make(http.Header)}
				next.ServeHTTP(rec, r)
				return &flight{code: rec.code, header: rec.header, body: rec.body}
			})
			for k, vals := range f.header {
				for _, v := range vals {
					w.Header().Add(k, v)
				}
			}
			w.WriteHeader(f.code)
			_, _ = w.Write(f.body)
		})
	}
}

// recorder captures a handler's response.
type recorder struct {
	code   int
	header http.Header
	body   []byte
}

func (r *recorder) Header() http.Header        { return r.header }
func (r *recorder) WriteHeader(code int)        { r.code = code }
func (r *recorder) Write(b []byte) (int, error) { r.body = append(r.body, b...); return len(b), nil }
