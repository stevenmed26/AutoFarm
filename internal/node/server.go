// internal/node/server.go
package node

import (
    "io"
    "log"
    "math/rand"
    "sync"
    "time"

    nodepb "github.com/stevenmed26/AutoFarm/internal/proto/nodepb"
    simulationpb "github.com/stevenmed26/AutoFarm/internal/proto/simulationpb"
)

type WorkerServer struct {
    nodepb.UnimplementedNodeWorkerServiceServer

    mu     sync.RWMutex
    states map[string]map[uint64]*simulationpb.EntityState
}

func NewWorkerServer() *WorkerServer {
    rand.Seed(time.Now().UnixNano())
    return &WorkerServer{
        states: make(map[string]map[uint64]*simulationpb.EntityState),
    }
}

// RunWorkerTicks implements a simple streaming worker:
// - Receives WorkerTickRequest messages from the orchestrator
// - Updates entity state in memory
// - Streams back WorkerTickResponse messages
func (s *WorkerServer) RunWorkerTicks(stream nodepb.NodeWorkerService_RunWorkerTicksServer) error {
    for {
        req, err := stream.Recv()
        if err == io.EOF {
            return nil
        }
        if err != nil {
            log.Printf("RunWorkerTicks recv error: %v", err)
            return err
        }

        start := time.Now()

        simID := req.GetSimulationId().GetValue()
        entityIDs := req.GetEntityIds()

        // Initialize state map for this simulation if needed.
        s.mu.Lock()
        simStates, ok := s.states[simID]
        if !ok {
            simStates = make(map[uint64]*simulationpb.EntityState)
            s.states[simID] = simStates
        }

        // Update entities.
        updated := make([]*simulationpb.EntityState, 0, len(entityIDs))

        for _, eid := range entityIDs {
            st, ok := simStates[eid]
            if !ok {
                st = s.newEntityState(eid)
                simStates[eid] = st
            }

            // Simple "simulation": move diagonally and drain battery.
            s.updateEntity(st)

            updated = append(updated, cloneEntityState(st))
        }
        s.mu.Unlock()

        computeMs := time.Since(start).Seconds() * 1000.0

        resp := &nodepb.WorkerTickResponse{
            SimulationId: req.GetSimulationId(),
            Tick:         req.GetTick(),
            Entities:     updated,
            ComputeMs:    computeMs,
        }

        if err := stream.Send(resp); err != nil {
            log.Printf("RunWorkerTicks send error: %v", err)
            return err
        }
    }
}

func (s *WorkerServer) newEntityState(id uint64) *simulationpb.EntityState {
    return &simulationpb.EntityState{
        EntityId: id,
        X:        rand.Float64() * 100,
        Y:        rand.Float64() * 100,
        Vx:       (rand.Float64() - 0.5) * 2, // -1 to +1
        Vy:       (rand.Float64() - 0.5) * 2,
        Battery:  100.0,
        Status:   "idle",
    }
}

func (s *WorkerServer) updateEntity(st *simulationpb.EntityState) {
    // Advance position.
    st.X += st.Vx
    st.Y += st.Vy

    // Simple boundary bounce.
    if st.X < 0 || st.X > 100 {
        st.Vx = -st.Vx
    }
    if st.Y < 0 || st.Y > 100 {
        st.Vy = -st.Vy
    }

    // Drain battery a bit.
    st.Battery -= 0.1
    if st.Battery < 0 {
        st.Battery = 0
        st.Status = "offline"
    } else if st.Battery < 20 {
        st.Status = "low_battery"
    } else {
        st.Status = "active"
    }
}

func cloneEntityState(st *simulationpb.EntityState) *simulationpb.EntityState {
    if st == nil {
        return nil
    }
    return &simulationpb.EntityState{
        EntityId: st.EntityId,
        X:        st.X,
        Y:        st.Y,
        Vx:       st.Vx,
        Vy:       st.Vy,
        Battery:  st.Battery,
        Status:   st.Status,
    }
}
