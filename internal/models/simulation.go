package models

// SimulationState represents the lifecycle state of a simulation.
type SimulationState string

const (
	SimulationStatePending SimulationState = "pending"
	SimulationStateRunning SimulationState = "running"
	SimulationStateStopped SimulationState = "stopped"
	SimulationStateFailed  SimulationState = "failed"
)

// Simulation describes a simulation tracked by the orchestrator.
type Simulation struct {
	ID          string
	Name        string
	Description string
	State       SimulationState
}
