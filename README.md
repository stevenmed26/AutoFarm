# AutoFarm — Distributed Simulation Platform
**Go • gRPC • WebSockets • Docker • AWS**  
**Real-time simulation backend with sub-100ms update latency**

AutoFarm is a distributed Go microservices system for running real-time simulations of autonomous farm robots.  
It uses gRPC for inter-service communication, WebSockets for live dashboard updates, and a scalable worker architecture for handling large simulation workloads.

---

## Overview

AutoFarm simulates large fleets of autonomous agents (robots, drones, tractors, etc.) across multiple distributed services.  
A central **Orchestrator** coordinates simulation ticks, distributes computation to **Node Workers**, aggregates results, and streams updates to clients through a WebSocket-enabled **API Gateway**.

The system is designed to demonstrate backend engineering patterns including microservices, low-latency pipelines, async worker pools, service boundaries, observability, and cloud-ready deployment.

---

## Architecture Summary

**Core Services**
- **API Gateway**  
  REST endpoints for simulation lifecycle. Hosts WebSocket connections for real-time updates.

- **Simulation Orchestrator**  
  Responsible for scheduling ticks, partitioning work, coordinating distributed workers, and aggregating results.

- **Node Worker Service**  
  Stateless worker microservice that processes subsets of simulation entities. Scales horizontally.

- **Data Layer**  
  PostgreSQL for simulation metadata. Optional Redis for state caching or pub/sub.

**Communication**
- gRPC for worker–orchestrator communication  
- WebSockets for client real-time updates  
- Internal channels or Redis for async event flow  

---

## Project Structure

```
autofarm/
  cmd/
    api/
    orchestrator/
    node/
  internal/
    api/
    orchestrator/
    node/
    proto/
    models/
    store/
    metrics/
    config/
  deployments/
    docker/
    k8s/
    terraform/
  scripts/
    loadtest/
  docs/
```

---

## How the System Works

1. **Simulation creation**  
   User submits a simulation configuration via REST.  
   The orchestrator initializes internal state and prepares distributed workers.

2. **Tick scheduling**  
   The orchestrator runs a timed loop (e.g., 20–60 ms per tick).  
   Each tick is dispatched to workers using gRPC streaming calls.

3. **Distributed computation**  
   Node workers process assigned entities in parallel, updating:
   - movement  
   - energy/battery  
   - physics  
   - task progress  

4. **Result aggregation**  
   Orchestrator receives partial updates from workers, merges them, and emits a unified simulation state.

5. **WebSocket broadcast**  
   API Gateway pushes updates to connected dashboards with end-to-end latency typically under 100ms.

---

## Running Locally

You can run the entire stack using Docker Compose.

```bash
docker-compose up --build
```

This starts:

- API Gateway (REST/WebSocket)
- Orchestrator
- Node workers (multiple replicas)
- PostgreSQL
- Redis (optional)

API default: `localhost:8080`

---

## Example Endpoints

```
POST /simulations
POST /simulations/{id}/start
POST /simulations/{id}/pause
POST /simulations/{id}/stop
GET  /simulations/{id}
GET  /ws/simulations/{id}
```

---

## gRPC Services (Conceptual)

Simulation tick service:

```proto
rpc RunSimulationTick(stream SimulationTickRequest)
  returns (stream SimulationTickResult);
```

Node compute service:

```proto
rpc ComputeTick(NodeComputeRequest)
  returns (NodeComputeResponse);
```

---

## Load Testing

Located in `scripts/loadtest/`.

Measures:

- tick execution latency  
- WebSocket latency  
- worker throughput  
- CPU/memory behavior  

---

## Deployment

Supports multiple deployment strategies:

- ECS or EKS for service orchestration  
- RDS for PostgreSQL  
- Optional Redis via ElastiCache  
- ALB for REST/WebSocket routing  
- Terraform modules for infra provisioning  
- GitHub Actions for CI/CD  

---

## Roadmap

- Additional simulation models  
- Predictive time-series outputs  
- Visual map UI  
- Prometheus/Grafana metrics  
- Auto-scaling optimizations  

---

## Purpose

AutoFarm is a portfolio-quality backend engineering project demonstrating:

- distributed systems  
- Go microservices  
- gRPC streaming  
- real-time WebSocket pipelines  
- async worker pools  
- cloud deployment  
- low-latency system design  
