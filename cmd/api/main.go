// cmd/api/main.go
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/stevenmed26/AutoFarm/internal/api"
	simulationpb "github.com/stevenmed26/AutoFarm/internal/proto/simulationpb"
)

func main() {
	httpAddr := getEnv("API_HTTP_ADDR", ":8080")
	orchestratorAddr := getEnv("ORCHESTRATOR_GRPC_ADDR", "localhost:50051")

	// Set up gRPC client to orchestrator.
	conn, err := grpc.Dial(
		orchestratorAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("failed to connect to orchestrator at %s: %v", orchestratorAddr, err)
	}
	defer conn.Close()

	simClient := simulationpb.NewSimulationServiceClient(conn)

	// Set up API server + WebSocket hub.
	server := api.NewServer(simClient)

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:         httpAddr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("API server listening on %s (orchestrator: %s)", httpAddr, orchestratorAddr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server failed: %v", err)
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
