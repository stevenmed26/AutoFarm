# AutoFarm Scaling Guide

## Overview
AutoFarm is built for distributed execution. Scaling is primarily accomplished through:
- Horizontal replication of Node Workers
- Distributed tick processing
- Cloud auto-scaling policies
- Efficient batching and delta updates

---

## Horizontal Scaling

### Node Workers
Each Worker handles a subset of simulation entities.  
Scaling Workers increases total compute throughput.

Example:
- 2 workers → 200 entities
- 4 workers → 400 entities
- 10 workers → 1000+ entities

The Orchestrator dynamically partitions workloads to available workers.

---

## Tick Rate Considerations

Higher tick rate = more real-time responsiveness but higher compute load.

Typical values:
- 20 ms (50 ticks/sec)
- 50 ms (20 ticks/sec)
- 100 ms (10 ticks/sec)

Scaling strategies:
- Reduce tick rate under heavy load
- Drop non-critical updates (client still sees smooth motion)

---

## Database Scaling

### PostgreSQL
Use:
- Connection pooling
- Read replicas (if metrics-heavy)
- Partitioned simulation history tables

### Redis
Cache frequently accessed state:
- Snapshot of latest simulation tick
- Worker registry metadata

---

## Load Balancing

### gRPC Load Balancing
ECS/EKS services expose gRPC endpoint for worker pool:
- Round-robin over workers
- Health-checking for misbehaving workers

### WebSocket Load Balancing
Use ALB in AWS for sticky sessions.

---

## Deployment Scaling Patterns

### ECS
Set auto-scaling policies:
- CPU > 60% → scale out worker tasks
- Memory > 70% → scale out worker tasks

### EKS
Use:
- Horizontal Pod Autoscalers
- Cluster Autoscaler
- PodDisruptionBudgets for safe rollouts

---

## Performance Optimization Techniques

- Pre-allocate entity arrays
- Reduce JSON size for WebSocket messages
- Enable binary/gzip compression
- Use delta updates instead of full-state payloads
- Tune gRPC window sizes
- Parallelize tick merge operations in the Orchestrator

---

Scaling should always maintain <100 ms update latency under target load.
