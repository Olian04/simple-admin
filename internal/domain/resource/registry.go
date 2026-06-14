package resource

import "sync"

// Registry holds the live set of resource runtimes grouped by backend. It is
// safe for concurrent use and atomically swappable, so a future discovery poll
// can replace the whole set without blocking readers.
type Registry struct {
	mu        sync.RWMutex
	order     []*Runtime          // stable insertion order
	byKey     map[string]*Runtime // "<backendID>/<name>"
	backends  []Backend           // first-seen order
	byBackend map[string][]*Runtime
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		byKey:     map[string]*Runtime{},
		byBackend: map[string][]*Runtime{},
	}
}

// Replace atomically swaps the registry contents.
func (r *Registry) Replace(runtimes []*Runtime) {
	order := make([]*Runtime, 0, len(runtimes))
	byKey := make(map[string]*Runtime, len(runtimes))
	byBackend := map[string][]*Runtime{}
	var backends []Backend
	seenBackend := map[string]bool{}

	for _, rt := range runtimes {
		if rt == nil {
			continue
		}
		key := rt.Descriptor.Key()
		bid := rt.Descriptor.Backend.ID
		order = append(order, rt)
		byKey[key] = rt
		byBackend[bid] = append(byBackend[bid], rt)
		if !seenBackend[bid] {
			seenBackend[bid] = true
			backends = append(backends, rt.Descriptor.Backend)
		}
	}

	r.mu.Lock()
	r.order = order
	r.byKey = byKey
	r.byBackend = byBackend
	r.backends = backends
	r.mu.Unlock()
}

// Lookup returns the runtime for a backend/resource pair.
func (r *Registry) Lookup(backendID, name string) (*Runtime, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rt, ok := r.byKey[backendID+"/"+name]
	return rt, ok
}

// Backends returns the backends in first-seen order.
func (r *Registry) Backends() []Backend {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Backend, len(r.backends))
	copy(out, r.backends)
	return out
}

// ResourcesFor returns the runtimes belonging to a backend, in order.
func (r *Registry) ResourcesFor(backendID string) []*Runtime {
	r.mu.RLock()
	defer r.mu.RUnlock()
	src := r.byBackend[backendID]
	out := make([]*Runtime, len(src))
	copy(out, src)
	return out
}

// First returns the first runtime in insertion order (used for the /ui redirect).
func (r *Registry) First() (*Runtime, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.order) == 0 {
		return nil, false
	}
	return r.order[0], true
}
