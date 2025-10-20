# AutoFarm
A small microservice cluster that simulates a fleet of autonomous “drones” doing farm tasks (patrol, scan, deliver). Each drone is an independent process. A Fleet Manager coordinates assignments and collects telemetry via gRPC bidirectional streams, then publishes live state to a WebSocket endpoint
