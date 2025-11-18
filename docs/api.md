# AutoFarm API Documentation

## Overview
The API Gateway exposes REST endpoints for simulation lifecycle operations 
and WebSocket endpoints for real-time simulation updates.

All responses are JSON unless otherwise noted.

---

# REST Endpoints

## Create Simulation
```
POST /simulations
```
### Request Body
```json
{
  "name": "Test Simulation",
  "entities": 200,
  "tick_rate_ms": 50
}
```

### Response
```json
{
  "id": "sim-1234",
  "status": "created"
}
```

---

## Start Simulation
```
POST /simulations/{id}/start
```
Response:
```json
{
  "id": "sim-1234",
  "status": "running"
}
```

---

## Pause Simulation
```
POST /simulations/{id}/pause
```
Response:
```json
{
  "id": "sim-1234",
  "status": "paused"
}
```

---

## Stop Simulation
```
POST /simulations/{id}/stop
```
Response:
```json
{
  "id": "sim-1234",
  "status": "stopped"
}
```

---

## Get Simulation Status
```
GET /simulations/{id}
```
Response:
```json
{
  "id": "sim-1234",
  "status": "running",
  "entities": 200,
  "tick_rate_ms": 50
}
```

---

# WebSocket Endpoints

## Subscribe to Simulation Updates
```
GET /ws/simulations/{id}
```

Client receives updates:
```json
{
  "simulation_id": "sim-1234",
  "tick": 148,
  "entities": [
    { "id": 1, "x": 10.2, "y": 3.1, "battery": 82.3 },
    { "id": 2, "x": 6.7, "y": 8.4, "battery": 77.9 }
  ]
}
```

---

# Error Codes

| Status Code | Meaning |
|-------------|---------|
| 400 | Invalid request |
| 404 | Simulation not found |
| 409 | Invalid lifecycle transition |
| 500 | Internal server error |

---

## gRPC Services (High-Level)

### Simulation Service
```proto
service SimulationOrchestrator {
    rpc RunSimulationTick (stream SimulationTickRequest)
        returns (stream SimulationTickResult);
}
```

### Worker Service
```proto
service NodeWorker {
    rpc ComputeTick (NodeComputeRequest)
        returns (NodeComputeResponse);
}
```

---

This API is designed to remain stable across versions, with new optional fields added non-breakingly.
