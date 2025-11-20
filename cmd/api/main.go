// cmd/api/main.go
package main

import (
	"log"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/stevenmed26/AutoFarm/internal/api"
	"github.com/stevenmed26/AutoFarm/internal/config"
	simulationpb "github.com/stevenmed26/AutoFarm/internal/proto/simulationpb"
)

func main() {
	cfg := config.LoadAPIConfig()


	// Set up gRPC client to orchestrator.
	conn, err := grpc.Dial(
		cfg.OrchestratorGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("failed to connect to orchestrator at %s: %v", cfg.OrchestratorGRPCAddr, err)
	}
	defer conn.Close()

	simClient := simulationpb.NewSimulationServiceClient(conn)

	// Set up API server + WebSocket hub.
	server := api.NewServer(simClient)

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("API server listening on %s (orchestrator: %s)", cfg.HTTPAddr, cfg.OrchestratorGRPCAddr)
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
