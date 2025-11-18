// internal/api/http.go
package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	commonpb "github.com/stevenmed26/AutoFarm/internal/proto/commonpb"
	simulationpb "github.com/stevenmed26/AutoFarm/internal/proto/simulationpb"
)

type createSimulationRequest struct {
	Name        string `json:"name"`
	EntityCount uint32 `json:"entities"`
	TickRateMs  uint32 `json:"tick_rate_ms"`
	Scenario    string `json:"scenario_type"`
}

type simulationResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	EntityCount uint32 `json:"entities"`
	TickRateMs  uint32 `json:"tick_rate_ms"`
	Scenario    string `json:"scenario_type"`
}

func (s *Server) handleSimulations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleCreateSimulation(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleCreateSimulation(w http.ResponseWriter, r *http.Request) {
	var reqBody createSimulationRequest
	if err := decodeJSONBody(r, &reqBody); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if reqBody.Name == "" || reqBody.EntityCount == 0 || reqBody.TickRateMs == 0 {
		http.Error(w, "name, entities, and tick_rate_ms are required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.simClient.CreateSimulation(ctx, &simulationpb.CreateSimulationRequest{
		Config: &simulationpb.SimulationConfig{
			Name:         reqBody.Name,
			EntityCount:  reqBody.EntityCount,
			TickRateMs:   reqBody.TickRateMs,
			ScenarioType: reqBody.Scenario,
		},
	})
	if err != nil {
		http.Error(w, "failed to create simulation: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, toSimulationResponse(resp.GetSimulation()))
}

func (s *Server) handleSimulationByID(w http.ResponseWriter, r *http.Request) {
	// Path format: /simulations/{id} or /simulations/{id}/action
	path := strings.TrimPrefix(r.URL.Path, "/simulations/")
	if path == "" {
		http.NotFound(w, r)
		return
	}
	parts := strings.Split(path, "/")
	id := parts[0]

	if len(parts) == 1 {
		// /simulations/{id}
		switch r.Method {
		case http.MethodGet:
			s.handleGetSimulation(w, r, id)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	// /simulations/{id}/{action}
	action := parts[1]
	switch action {
	case "start":
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.handleStartSimulation(w, r, id)
	case "pause":
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.handlePauseSimulation(w, r, id)
	case "stop":
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.handleStopSimulation(w, r, id)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleGetSimulation(w http.ResponseWriter, r *http.Request, id string) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.simClient.GetSimulation(ctx, &simulationpb.GetSimulationRequest{
		Id: &commonpb.SimulationId{Value: id},
	})
	if err != nil {
		http.Error(w, "failed to get simulation: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, toSimulationResponse(resp.GetSimulation()))
}

func (s *Server) handleStartSimulation(w http.ResponseWriter, r *http.Request, id string) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.simClient.StartSimulation(ctx, &simulationpb.StartSimulationRequest{
		Id: &commonpb.SimulationId{Value: id},
	})
	if err != nil {
		http.Error(w, "failed to start simulation: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, toSimulationResponse(resp.GetSimulation()))
}

func (s *Server) handlePauseSimulation(w http.ResponseWriter, r *http.Request, id string) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.simClient.PauseSimulation(ctx, &simulationpb.PauseSimulationRequest{
		Id: &commonpb.SimulationId{Value: id},
	})
	if err != nil {
		http.Error(w, "failed to pause simulation: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, toSimulationResponse(resp.GetSimulation()))
}

func (s *Server) handleStopSimulation(w http.ResponseWriter, r *http.Request, id string) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.simClient.StopSimulation(ctx, &simulationpb.StopSimulationRequest{
		Id: &commonpb.SimulationId{Value: id},
	})
	if err != nil {
		http.Error(w, "failed to stop simulation: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, toSimulationResponse(resp.GetSimulation()))
}

func toSimulationResponse(sim *simulationpb.Simulation) *simulationResponse {
	if sim == nil || sim.Config == nil {
		return nil
	}
	return &simulationResponse{
		ID:          sim.GetId().GetValue(),
		Name:        sim.Config.GetName(),
		Status:      sim.GetStatus().String(),
		EntityCount: sim.Config.GetEntityCount(),
		TickRateMs:  sim.Config.GetTickRateMs(),
		Scenario:    sim.Config.GetScenarioType(),
	}
}

// helpers

func decodeJSONBody(r *http.Request, dst any) error {
	if r.Body == nil {
		return errors.New("empty body")
	}
	defer r.Body.Close()

	data, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1MB limit
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return errors.New("empty body")
	}
	return json.Unmarshal(data, dst)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}
