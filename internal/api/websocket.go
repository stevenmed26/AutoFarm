// internal/api/websocket.go
package api

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "strings"
    "time"

    "github.com/gorilla/websocket"

    commonpb "github.com/stevenmed26/AutoFarm/internal/proto/commonpb"
    simulationpb "github.com/stevenmed26/AutoFarm/internal/proto/simulationpb"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool { return true }, // relax for dev
}

// DashboardUpdate is the JSON payload shape sent to clients.
type DashboardUpdate struct {
    SimulationID string            `json:"simulation_id"`
    Tick         uint64            `json:"tick"`
    Entities     []DashboardEntity `json:"entities"`
    AvgComputeMs float64           `json:"avg_compute_ms"`
    WorkerCount  uint32            `json:"worker_count"`
    CompletedAt  time.Time         `json:"completed_at"`
}

type DashboardEntity struct {
    ID      uint64  `json:"id"`
    X       float64 `json:"x"`
    Y       float64 `json:"y"`
    Vx      float64 `json:"vx"`
    Vy      float64 `json:"vy"`
    Battery float64 `json:"battery"`
    Status  string  `json:"status"`
}

func dashboardUpdateFromProto(tick *simulationpb.AggregatedTick) *DashboardUpdate {
    entities := make([]DashboardEntity, 0, len(tick.GetEntities()))
    for _, e := range tick.GetEntities() {
        entities = append(entities, DashboardEntity{
            ID:      e.GetEntityId(),
            X:       e.GetX(),
            Y:       e.GetY(),
            Vx:      e.GetVx(),
            Vy:      e.GetVy(),
            Battery: e.GetBattery(),
            Status:  e.GetStatus(),
        })
    }

    var completedAt time.Time
    if ts := tick.GetCompletedAt(); ts != nil {
        completedAt = ts.AsTime()
    }

    return &DashboardUpdate{
        SimulationID: tick.GetSimulationId().GetValue(),
        Tick:         tick.GetTick(),
        Entities:     entities,
        AvgComputeMs: tick.GetAvgComputeMs(),
        WorkerCount:  tick.GetWorkerCount(),
        CompletedAt:  completedAt,
    }
}

// HTTP handler for WebSocket endpoint: /ws/simulations/{id}
func (s *Server) handleSimulationWebSocket(w http.ResponseWriter, r *http.Request) {
    path := strings.TrimPrefix(r.URL.Path, "/ws/simulations/")
    if path == "" {
        http.Error(w, "missing simulation id", http.StatusBadRequest)
        return
    }
    simID := strings.SplitN(path, "/", 2)[0]

    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("ws upgrade error: %v", err)
        return
    }
    defer conn.Close()

    // Fire a goroutine to watch client disconnects (optional, mostly to read pings).
    go func() {
        conn.SetReadLimit(1024)
        _ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
        conn.SetPongHandler(func(string) error {
            _ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
            return nil
        })
        for {
            if _, _, err := conn.ReadMessage(); err != nil {
                // client closed or error
                return
            }
        }
    }()

    ctx, cancel := context.WithCancel(r.Context())
    defer cancel()

    // Open a streaming RPC to the orchestrator.
    stream, err := s.simClient.StreamAggregatedTicks(ctx, &simulationpb.StreamAggregatedTicksRequest{
        Id: &commonpb.SimulationId{Value: simID},
    })
    if err != nil {
        log.Printf("StreamAggregatedTicks error for %s: %v", simID, err)
        return
    }

    pingTicker := time.NewTicker(30 * time.Second)
    defer pingTicker.Stop()

    for {
        // Pull a tick from gRPC or send pings.
        select {
        case <-ctx.Done():
            return
        case <-pingTicker.C:
            _ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        default:
            // Non-blocking check for next tick
        }

        tick, err := stream.Recv()
        if err != nil {
            log.Printf("StreamAggregatedTicks recv error for %s: %v", simID, err)
            return
        }

        update := dashboardUpdateFromProto(tick)
        data, err := json.Marshal(update)
        if err != nil {
            log.Printf("marshal dashboard update error: %v", err)
            continue
        }

        _ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
        if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
            return
        }
    }
}
