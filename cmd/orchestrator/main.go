// cmd/orchestrator/main.go
package main

import (
    "log"
    "net"

    "google.golang.org/grpc"

    "github.com/stevenmed26/AutoFarm/internal/orchestrator"
    "github.com/stevenmed26/AutoFarm/internal/config"
    "github.com/stevenmed26/AutoFarm/internal/store"
    simulationpb "github.com/stevenmed26/AutoFarm/internal/proto/simulationpb"
)

func main() {
    cfg := config.LoadOrchestratorConfig()

    db, err := store.ConnectWithRetry(cfg.DBDSN)
    if err != nil {
        log.Fatalf("failed to connect to database: %v", err)
    }
    defer db.Close()

    simStore := store.NewSimulationStore(db)

    lis, err := net.Listen("tcp", cfg.GRPCAddr)
    if err != nil {
        log.Fatalf("failed to listen on %s: %v", cfg.GRPCAddr, err)
    }

    grpcServer := grpc.NewServer()

    simServer := orchestrator.NewSimulationServer(cfg.WorkerGRPCAddr, simStore)
    simulationpb.RegisterSimulationServiceServer(grpcServer, simServer)

    log.Printf("Orchestrator gRPC server listening on %s (worker: %s)", cfg.GRPCAddr, cfg.WorkerGRPCAddr)

    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("failed to serve gRPC: %v", err)
    }
}
