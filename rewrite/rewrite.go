// Package rewrite provides HTTP request URL rewriting middleware.
package rewrite

import (
	"net/http"
	"regexp"
	"strings"
)

// Rule defines a single rewrite rule with a pattern and replacement.
type Rule struct {
	Pattern     *regexp.Regexp
	Replacement string
	Redirect    bool // if true, send an HTTP redirect instead of rewriting internally
}

// Options holds configuration for the rewrite middleware.
type Options struct {
	Rules []Rule
}

// DefaultOptions returns an Options with no rules.
func DefaultOptions() Options {
	return Options{}
}

// AddRule compiles a regex pattern and appends a rewrite rule to the options.
// If redirect is true the client receives a 301 redirect; otherwise the request
// path is rewritten transparently.
func (o *Options) AddRule(pattern, replacement string, redirect bool) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	o.Rules = append(o.Rules, Rule{
		Pattern:     re,
		Replacement: replacement,
		Redirect:    redirect,
	})
	return nil
}

// New returns middleware that rewrites or redirects request URLs based on the
// configured rules. Rules are evaluated in order; the first match wins.
func New(opts Options) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			for _, rule := range opts.Rules {
				if !rule.Pattern.MatchString(path) {
					continue
				}

				newPath := rule.Pattern.ReplaceAllString(path, rule.Replacement)

				if rule.Redirect {
					if r.URL.RawQuery != "" {
						newPath += "?" + r.URL.RawQuery
					}
					http.Redirect(w, r, newPath, http.StatusMovedPermanently)
					return
				}

				// Internal rewrite — mutate a shallow copy of the URL.
				r2 := r.Clone(r.Context())
				r2.URL.Path = newPath
				if !strings.HasPrefix(newPath, "/") {
					r2.URL.Path = "/" + newPath
				}
				r2.RequestURI = r2.URL.RequestURI()
				next.ServeHTTP(w, r2)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
