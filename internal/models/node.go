package models

// NodeStatus represents the lifecycle status of a node.
type NodeStatus string

const (
	NodeStatusUnknown NodeStatus = "unknown"
	NodeStatusOnline  NodeStatus = "online"
	NodeStatusOffline NodeStatus = "offline"
)

// Node describes a worker node participating in the AutoFarm cluster.
type Node struct {
	ID      string
	Address string
	Status  NodeStatus
}
