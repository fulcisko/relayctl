package upstream

import (
	"net/http"
	"strings"
)

// RewriteRule defines a path prefix rewrite: requests matching Prefix
// have that prefix stripped and replaced with Replacement.
type RewriteRule struct {
	Prefix      string
	Replacement string
}

// RewriteRegistry holds per-backend path rewrite rules.
type RewriteRegistry struct {
	rules map[string]RewriteRule
}

// NewRewriteRegistry returns an empty RewriteRegistry.
func NewRewriteRegistry() *RewriteRegistry {
	return &RewriteRegistry{rules: make(map[string]RewriteRule)}
}

// Set stores a rewrite rule for the given backend URL.
func (r *RewriteRegistry) Set(backend, prefix, replacement string) error {
	if backend == "" {
		return errEmptyBackend
	}
	r.rules[backend] = RewriteRule{Prefix: prefix, Replacement: replacement}
	return nil
}

// Get returns the rewrite rule for a backend, and whether one exists.
func (r *RewriteRegistry) Get(backend string) (RewriteRule, bool) {
	rule, ok := r.rules[backend]
	return rule, ok
}

// Delete removes the rewrite rule for a backend.
func (r *RewriteRegistry) Delete(backend string) {
	delete(r.rules, backend)
}

// Snapshot returns a copy of all current rewrite rules.
func (r *RewriteRegistry) Snapshot() map[string]RewriteRule {
	out := make(map[string]RewriteRule, len(r.rules))
	for k, v := range r.rules {
		out[k] = v
	}
	return out
}

// Apply rewrites the request path for the given backend if a rule exists.
func (r *RewriteRegistry) Apply(backend string, req *http.Request) {
	rule, ok := r.rules[backend]
	if !ok {
		return
	}
	if rule.Prefix != "" && strings.HasPrefix(req.URL.Path, rule.Prefix) {
		req.URL.Path = rule.Replacement + strings.TrimPrefix(req.URL.Path, rule.Prefix)
		if req.URL.RawPath != "" {
			req.URL.RawPath = rule.Replacement + strings.TrimPrefix(req.URL.RawPath, rule.Prefix)
		}
	}
}

// Len returns the number of rewrite rules currently registered.
func (r *RewriteRegistry) Len() int {
	return len(r.rules)
}

// errEmptyBackend is a sentinel for missing backend keys.
var errEmptyBackend = errRewriteEmptyBackend("backend must not be empty")

type errRewriteEmptyBackend string

func (e errRewriteEmptyBackend) Error() string { return string(e) }
