# AutoFarm Architecture

## Overview
AutoFarm is a distributed simulation backend composed of independently deployable Go microservices. 
Each service is responsible for a specific subset of the system’s functionality: API Gateway, Simulation Orchestrator, 
Node Worker services, WebSocket broadcaster, and supporting data stores.

---

## High-Level Architecture

AutoFarm operates on a tick-based simulation model:
1. A user initiates a simulation through the API Gateway.
2. The Orchestrator schedules ticks at a fixed interval (20–60 ms).
3. Simulation work is partitioned and sent to Node Workers using gRPC streaming.
4. Workers compute state updates for their assigned entities.
5. Orchestrator aggregates worker results.
6. Aggregated results are published to the WebSocket broadcaster.
7. Dashboard clients receive updates with sub-100ms end-to-end latency.

---

## Core Services

### API Gateway
- Provides REST endpoints for simulation lifecycle management.
- Hosts WebSocket connections and handles client subscriptions.
- Forwards simulation control commands to the Orchestrator.
- Emits aggregated simulation data to subscribed WebSocket clients.

### Simulation Orchestrator
- Manages simulation lifecycle and tick scheduling.
- Partitions entity workloads across Node Workers.
- Maintains worker registry and load distribution logic.
- Aggregates responses from workers before broadcasting.

### Node Worker Service
- Stateless microservice that processes simulation entities.
- Contains a goroutine worker pool for parallel computation.
- Implements physics, state transitions, battery/energy modeling, etc.
- Designed to scale horizontally under load.

### WebSocket Broadcaster
- Fan-out service for real-time updates.
- Maintains client subscription maps per simulation.
- Uses async, non-blocking send patterns.

---

## Data Layer

### PostgreSQL
Stores:
- Simulation configurations
- Entity metadata
- Simulation history
- Aggregated metrics snapshots

### Redis (optional)
Provides:
- In-memory caching
- Pub/sub for internal event propagation
- Distributed lock functionality (if needed)

---

## Internal Communication

### gRPC (Orchestrator <-> Node Workers)
Chosen for:
- Binary encoding efficiency
- Bi-directional streaming support
- Low-latency RPC calls

Key RPCs:
- `RunSimulationTick(stream SimulationTickRequest) → stream SimulationTickResult`
- `ComputeTick(NodeComputeRequest) → NodeComputeResponse`

### WebSockets (API Gateway <-> Clients)
Used for low-latency dashboard updates.

---

## Deployment Strategy

### Local (Docker Compose)
Services run as isolated containers:
- API (HTTP/WebSocket)
- Orchestrator
- Worker replicas
- PostgreSQL
- Redis (optional)

### Production (AWS)
Supports:
- ECS/EKS deployment
- ALB routing
- RDS PostgreSQL
- ElastiCache Redis
- Horizontal auto-scaling of Node Workers

---

## Observability

AutoFarm includes:
- Structured JSON logging
- Per-tick latency metrics
- Worker throughput counters
- WebSocket broadcast timing
- Optional Prometheus exporters

