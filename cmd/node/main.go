// cmd/node/main.go
package main

import (
    "log"
    "net"

    "google.golang.org/grpc"

    "github.com/stevenmed26/AutoFarm/internal/node"
    nodepb "github.com/stevenmed26/AutoFarm/internal/proto/nodepb"
)

func main() {
    addr := ":50052"

    lis, err := net.Listen("tcp", addr)
    if err != nil {
        log.Fatalf("failed to listen on %s: %v", addr, err)
    }

    grpcServer := grpc.NewServer()

    workerServer := node.NewWorkerServer()
    nodepb.RegisterNodeWorkerServiceServer(grpcServer, workerServer)

    log.Printf("Node Worker gRPC server listening on %s", addr)

    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("failed to serve gRPC: %v", err)
    }
}
