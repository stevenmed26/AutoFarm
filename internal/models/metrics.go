package models

// NodeMetrics represents basic runtime stats for a node.
type NodeMetrics struct {
	NodeID           string
	CPUPercent       float64
	MemoryBytes      uint64
	ActiveSimulations int
}

// SimulationMetrics represents high-level statistics about a simulation.
type SimulationMetrics struct {
	SimulationID   string
	TicksProcessed uint64
	EntitiesActive int
}
