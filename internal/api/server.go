// internal/api/server.go
package api

import (
	"net/http"

	simulationpb "github.com/stevenmed26/AutoFarm/internal/proto/simulationpb"
)

type Server struct {
	simClient simulationpb.SimulationServiceClient
}

func NewServer(simClient simulationpb.SimulationServiceClient) *Server {
	return &Server{
		simClient: simClient,
	}
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	// REST API
	mux.HandleFunc("/simulations", s.handleSimulations)
	mux.HandleFunc("/simulations/", s.handleSimulationByID)

	// WebSocket stream for dashboard
	mux.HandleFunc("/ws/simulations/", s.handleSimulationWebSocket)

	// Health check
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Static dashboard (served from ./web)
	// This should be last so more specific paths like /simulations don't get shadowed.
	fileServer := http.FileServer(http.Dir("web"))
	mux.Handle("/", fileServer)
}

