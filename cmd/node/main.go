// cmd/node/main.go
package main

import (
    "log"
    "net"

    "google.golang.org/grpc"

    "github.com/stevenmed26/AutoFarm/internal/node"
    "github.com/stevenmed26/AutoFarm/internal/config"
    nodepb "github.com/stevenmed26/AutoFarm/internal/proto/nodepb"
)

func main() {
    cfg := config.LoadNodeConfig()

    lis, err := net.Listen("tcp", cfg.GRPCAddr)
    if err != nil {
        log.Fatalf("failed to listen on %s: %v", cfg.GRPCAddr, err)
    }

    grpcServer := grpc.NewServer()

    workerServer := node.NewWorkerServer()
    nodepb.RegisterNodeWorkerServiceServer(grpcServer, workerServer)

    log.Printf("Node Worker gRPC server listening on %s", cfg.GRPCAddr)

    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("failed to serve gRPC: %v", err)
    }
}
