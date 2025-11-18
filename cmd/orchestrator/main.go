// cmd/orchestrator/main.go
package main

import (
    "log"
    "net"

    "google.golang.org/grpc"

    "github.com/stevenmed26/AutoFarm/internal/orchestrator"
    simulationpb "github.com/stevenmed26/AutoFarm/internal/proto/simulationpb"
)

func main() {
    addr := ":50051"

    lis, err := net.Listen("tcp", addr)
    if err != nil {
        log.Fatalf("failed to listen on %s: %v", addr, err)
    }

    grpcServer := grpc.NewServer()

    simServer := orchestrator.NewSimulationServer()
    simulationpb.RegisterSimulationServiceServer(grpcServer, simServer)

    log.Printf("Orchestrator gRPC server listening on %s", addr)

    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("failed to serve gRPC: %v", err)
    }
}
