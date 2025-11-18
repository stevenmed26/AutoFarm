// internal/api/server.go
package api

import (
	"net/http"

	simulationpb "github.com/stevenmed26/AutoFarm/internal/proto/simulationpb"
)

type Server struct {
	simClient simulationpb.SimulationServiceClient
	hub       *Hub
}

func NewServer(simClient simulationpb.SimulationServiceClient) *Server {
	return &Server{
		simClient: simClient,
		hub:       NewHub(),
	}
}

// RegisterRoutes wires up all HTTP and WebSocket endpoints.
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	// REST
	mux.HandleFunc("/simulations", s.handleSimulations)                  // POST /simulations
	mux.HandleFunc("/simulations/", s.handleSimulationByID)              // POST/GET /simulations/{id}/...

	// WebSocket
	mux.HandleFunc("/ws/simulations/", s.handleSimulationWebSocket)      // GET /ws/simulations/{id}

	// Optionally: health check
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}
