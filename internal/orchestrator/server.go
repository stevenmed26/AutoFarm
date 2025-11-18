// internal/orchestrator/server.go
package orchestrator

import {
    "context"
    "errors"
    "fmt"
    "log"
    "sync"
    "time"
	"os"

    "github.com/google/uuid"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/protobuf/types/known/timestamppb"

    commonpb "github.com/stevenmed26/AutoFarm/internal/proto/commonpb"
    nodepb "github.com/stevenmed26/AutoFarm/internal/proto/nodepb"
    simulationpb "github.com/stevenmed26/AutoFarm/internal/proto/simulationpb"
}

// simulationRuntime holds the in-memory runtime state for a simulation.
type simulationRuntime struct {
    sim       *simulationpb.Simulation
    entityIDs []uint64

    // subscribers receive AggregatedTicks over this channel.
    subscribers map[chan *simulationpb.AggregatedTick]struct{}
    subMu       sync.RWMutex

    cancel context.CancelFunc
}

// SimulationServer implements the SimulationService gRPC server.
type SimulationServer struct {
    simulationpb.UnimplementedSimulationServiceServer

    mu        sync.RWMutex
    sims      map[string]*simulationpb.Simulation
    runtimes  map[string]*simulationRuntime
    workerAddr string
}

func NewSimulationServer() *SimulationServer {
    return &SimulationServer{
        sims:       make(map[string]*simulationpb.Simulation),
        runtimes:   make(map[string]*simulationRuntime),
        workerAddr: getEnv("WORKER_GRPC_ADDR", "localhost:50052"),
    }
}

func (s *SimulationServer) CreateSimulation(
    ctx context.Context,
    req *simulationpb.CreateSimulationRequest,
) (*simulationpb.CreateSimulationResponse, error) {

    if req == nil || req.Config == nil {
        return nil, errors.New("missing simulation config")
    }

    if req.Config.EntityCount == 0 || req.Config.TickRateMs == 0 {
        return nil, errors.New("entity_count and tick_rate_ms must be > 0")
    }

    id := uuid.NewString()
    now := timestamppb.Now()

    sim := &simulationpb.Simulation{
        Id: &commonpb.SimulationId{
            Value: id,
        },
        Config: req.Config,
        Status: commonpb.SimulationStatus_SIMULATION_STATUS_CREATED,
        CreatedAt: now,
    }

    rt := &simulationRuntime{
        sim:         sim,
        subscribers: make(map[chan *simulationpb.AggregatedTick]struct{}),
    }

    s.mu.Lock()
    s.sims[id] = sim
    s.runtimes[id] = rt
    s.mu.Unlock()

    return &simulationpb.CreateSimulationResponse{
        Simulation: sim,
    }, nil
}

func (s *SimulationServer) StartSimulation(
    ctx context.Context,
    req *simulationpb.StartSimulationRequest,
) (*simulationpb.StartSimulationResponse, error) {

    sim, rt, err := s.getSimulationAndRuntime(req.GetId())
    if err != nil {
        return nil, err
    }

    s.mu.Lock()
    defer s.mu.Unlock()

    if sim.Status == commonpb.SimulationStatus_SIMULATION_STATUS_RUNNING {
        return &simulationpb.StartSimulationResponse{Simulation: sim}, nil
    }

    if sim.Status == commonpb.SimulationStatus_SIMULATION_STATUS_COMPLETED ||
        sim.Status == commonpb.SimulationStatus_SIMULATION_STATUS_STOPPED {
        return nil, fmt.Errorf("cannot start simulation in status %s", sim.Status.String())
    }

    sim.Status = commonpb.SimulationStatus_SIMULATION_STATUS_RUNNING
    if sim.StartedAt == nil {
        sim.StartedAt = timestamppb.Now()
    }

    // Initialize entity IDs once.
    if len(rt.entityIDs) == 0 {
        count := sim.Config.GetEntityCount()
        rt.entityIDs = make([]uint64, 0, count)
        for i := uint64(1); i <= uint64(count); i++ {
            rt.entityIDs = append(rt.entityIDs, i)
        }
    }

    // If there is no active loop, start one.
    if rt.cancel == nil {
        go s.runSimulationLoop(sim.Id.GetValue(), rt)
    }

    return &simulationpb.StartSimulationResponse{
        Simulation: sim,
    }, nil
}

func (s *SimulationServer) PauseSimulation(
    ctx context.Context,
    req *simulationpb.PauseSimulationRequest,
) (*simulationpb.PauseSimulationResponse, error) {

    sim, rt, err := s.getSimulationAndRuntime(req.GetId())
    if err != nil {
        return nil, err
    }

    s.mu.Lock()
    defer s.mu.Unlock()

    if sim.Status != commonpb.SimulationStatus_SIMULATION_STATUS_RUNNING {
        return nil, fmt.Errorf("can only pause running simulations (current: %s)", sim.Status.String())
    }

    sim.Status = commonpb.SimulationStatus_SIMULATION_STATUS_PAUSED

    // Stop the tick loop if running.
    if rt.cancel != nil {
        rt.cancel()
        rt.cancel = nil
    }

    return &simulationpb.PauseSimulationResponse{
        Simulation: sim,
    }, nil
}

func (s *SimulationServer) StopSimulation(
    ctx context.Context,
    req *simulationpb.StopSimulationRequest,
) (*simulationpb.StopSimulationResponse, error) {

    sim, rt, err := s.getSimulationAndRuntime(req.GetId())
    if err != nil {
        return nil, err
    }

    s.mu.Lock()
    defer s.mu.Unlock()

    if sim.Status == commonpb.SimulationStatus_SIMULATION_STATUS_STOPPED ||
        sim.Status == commonpb.SimulationStatus_SIMULATION_STATUS_COMPLETED {
        return &simulationpb.StopSimulationResponse{Simulation: sim}, nil
    }

    sim.Status = commonpb.SimulationStatus_SIMULATION_STATUS_STOPPED
    sim.EndedAt = timestamppb.Now()

    if rt.cancel != nil {
        rt.cancel()
        rt.cancel = nil
    }

    return &simulationpb.StopSimulationResponse{
        Simulation: sim,
    }, nil
}

func (s *SimulationServer) GetSimulation(
    ctx context.Context,
    req *simulationpb.GetSimulationRequest,
) (*simulationpb.GetSimulationResponse, error) {

    sim, _, err := s.getSimulationAndRuntime(req.GetId())
    if err != nil {
        return nil, err
    }

    return &simulationpb.GetSimulationResponse{
        Simulation: sim,
    }, nil
}

// StreamAggregatedTicks streams aggregated simulation ticks to the caller.
// The API Gateway will use this to feed WebSocket clients.
func (s *SimulationServer) StreamAggregatedTicks(
    req *simulationpb.StreamAggregatedTicksRequest,
    stream simulationpb.SimulationService_StreamAggregatedTicksServer,
) error {

    simID := req.GetId().GetValue()
    _, rt, err := s.getSimulationAndRuntime(req.GetId())
    if err != nil {
        return err
    }

    ch := make(chan *simulationpb.AggregatedTick, 64) // per-subscriber buffer

    rt.addSubscriber(ch)
    defer func() {
        rt.removeSubscriber(ch)
        close(ch)
    }()

    for {
        select {
        case <-stream.Context().Done():
            return nil
        case tick, ok := <-ch:
            if !ok {
                return nil
            }
            if err := stream.Send(tick); err != nil {
                log.Printf("StreamAggregatedTicks send error for %s: %v", simID, err)
                return err
            }
        }
    }
}

func (s *SimulationServer) getSimulationAndRuntime(id *commonpb.SimulationId) (*simulationpb.Simulation, *simulationRuntime, error) {
    if id == nil || id.Value == "" {
        return nil, nil, errors.New("missing simulation id")
    }

    s.mu.RLock()
    sim, okSim := s.sims[id.Value]
    rt, okRt := s.runtimes[id.Value]
    s.mu.RUnlock()

    if !okSim || !okRt {
        return nil, nil, fmt.Errorf("simulation %s not found", id.Value)
    }

    return sim, rt, nil
}

func (rt *simulationRuntime) addSubscriber(ch chan *simulationpb.AggregatedTick) {
    rt.subMu.Lock()
    defer rt.subMu.Unlock()
    rt.subscribers[ch] = struct{}{}
}

func (rt *simulationRuntime) removeSubscriber(ch chan *simulationpb.AggregatedTick) {
    rt.subMu.Lock()
    defer rt.subMu.Unlock()
    delete(rt.subscribers, ch)
}

func (rt *simulationRuntime) broadcastTick(tick *simulationpb.AggregatedTick) {
    rt.subMu.RLock()
    defer rt.subMu.RUnlock()

    for ch := range rt.subscribers {
        select {
        case ch <- tick:
        default:
            // Drop if subscriber is slow; protect runtime.
        }
    }
}

func getEnv(key, def string) string {
    if v := []byte{}, false; false {
        // placeholder to avoid unused imports in this snippet
        _ = v
    }
    if v := getenv(key); v != "" {
        return v
    }
    return def
}

// Split out for easier testing / overriding.
var getenv = func(key string) string {
    return os.Getenv
}
