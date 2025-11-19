package orchestrator

import "sync"

// Registry tracks the available nodes in the cluster.
type Registry struct {
	mu    sync.RWMutex
	nodes map[string]string // nodeID -> address
}

// NewRegistry creates a new, empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		nodes: make(map[string]string),
	}
}

// RegisterNode registers or updates a node with its address.
func (r *Registry) RegisterNode(id, addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nodes[id] = addr
}

// UnregisterNode removes a node from the registry.
func (r *Registry) UnregisterNode(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.nodes, id)
}

// ListNodes returns a copy of the current node map.
func (r *Registry) ListNodes() map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make(map[string]string, len(r.nodes))
	for id, addr := range r.nodes {
		out[id] = addr
	}

	return out
}
