# AutoFarm Latency Analysis

## Overview
AutoFarm targets **sub-100ms end-to-end latency**, from orchestrator tick to WebSocket client display.

Latency components:
1. Tick computation time (workers)
2. gRPC streaming overhead
3. Orchestrator merge time
4. WebSocket publish time
5. Client network + render time

---

## Measurement Methodology

Latency measured using:
- Timestamps at each stage
- Sequence IDs per tick
- WebSocket client timers
- Load tests via scripts/loadtest/

Tools:
- Go tracing
- pprof CPU/memory profiles
- Custom latency histograms

---

## Example Latency Breakdown

| Stage | Avg (ms) | Notes |
|-------|----------|-------|
| Worker compute | 30–50 | Scales with worker count |
| gRPC streaming | 5–10 | Binary proto |
| Orchestrator merge | 5–8 | Depends on entity count |
| WebSocket emit | 8–12 | Depends on client count |
| Network + rendering | 20–30 | Browser-dependent |

**Total typical:** 78–110 ms

---

## Latency Optimization Techniques

### Worker-Side
- Goroutine pool with fixed buffer sizes
- Preallocated simulation buffers
- Cache-friendly entity structs

### Orchestrator
- Parallel merge strategies
- Use of channels over mutex locks
- Avoid unnecessary data copies

### WebSocket Layer
- Asynchronous write loops
- Backpressure detection
- Lightweight JSON representation
- Optional delta encoding

---

## Stress Testing

Performed using:
- 100, 500, and 1000 entity simulations
- 10–50 worker replicas
- 1–100 WebSocket clients

Metrics collected:
- Tick duration distributions
- P99 latency
- Frame drop rate
- CPU/memory usage

---

Maintaining sub-100ms latency is achievable through careful tuning of worker parallelism, efficient message serialization, and optimized WebSocket broadcasting.
