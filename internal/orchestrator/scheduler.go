package orchestrator

// Scheduler is responsible for picking which node should handle work.
type Scheduler struct {
	registry *Registry
}

// NewScheduler creates a new Scheduler using the given Registry.
func NewScheduler(r *Registry) *Scheduler {
	return &Scheduler{
		registry: r,
	}
}

// ChooseNode returns the ID of a node that can handle work.
// This is a placeholder: it just returns the first node it sees.
func (s *Scheduler) ChooseNode() string {
	nodes := s.registry.ListNodes()
	for id := range nodes {
		return id
	}
	return ""
}
