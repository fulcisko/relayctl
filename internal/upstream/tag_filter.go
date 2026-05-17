package upstream

import (
	"net/http"
	"slices"
)

// TagFilterBalancer wraps an inner Balancer and restricts candidates to
// backends that carry a specific tag. If none match, it falls back to the
// inner balancer so requests are never dropped.
type TagFilterBalancer struct {
	inner    Balancer
	tags     *TagRegistry
	required string
}

// NewTagFilterBalancer returns a TagFilterBalancer that only picks backends
// tagged with required. inner and tags must be non-nil.
func NewTagFilterBalancer(inner Balancer, tags *TagRegistry, required string) *TagFilterBalancer {
	if inner == nil {
		panic("tag_filter: inner balancer must not be nil")
	}
	if tags == nil {
		panic("tag_filter: tag registry must not be nil")
	}
	return &TagFilterBalancer{inner: inner, tags: tags, required: required}
}

// Next returns the next backend that carries the required tag. If no backend
// in the inner balancer's pool is tagged, the inner balancer decides.
func (f *TagFilterBalancer) Next(r *http.Request) (string, func()) {
	matched := f.filtered()
	if len(matched) == 0 {
		return f.inner.Next(r)
	}
	// Round-robin over matched subset via inner balancer until we land on one.
	for range len(f.inner.Backends()) + 1 {
		backend, done := f.inner.Next(r)
		if slices.Contains(matched, backend) {
			return backend, done
		}
		if done != nil {
			done()
		}
	}
	// Fallback: return whatever the inner balancer gives us.
	return f.inner.Next(r)
}

// Backends returns the full backend list from the inner balancer.
func (f *TagFilterBalancer) Backends() []string {
	return f.inner.Backends()
}

// RequiredTag returns the tag this balancer filters on.
func (f *TagFilterBalancer) RequiredTag() string {
	return f.required
}

// SetRequiredTag updates the tag filter at runtime.
func (f *TagFilterBalancer) SetRequiredTag(tag string) {
	f.required = tag
}

func (f *TagFilterBalancer) filtered() []string {
	all := f.inner.Backends()
	out := make([]string, 0, len(all))
	for _, b := range all {
		if ts, ok := f.tags.Get(b); ok && slices.Contains(ts, f.required) {
			out = append(out, b)
		}
	}
	return out
}
