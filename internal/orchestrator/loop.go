// internal/orchestrator/loop.go
package orchestrator

import (
    "context"
    "log"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/protobuf/types/known/timestamppb"

    nodepb "github.com/stevenmed26/AutoFarm/internal/proto/nodepb"
    simulationpb "github.com/stevenmed26/AutoFarm/internal/proto/simulationpb"
)

func (s *SimulationServer) runSimulationLoop(simID string, rt *simulationRuntime) {
    // Create a dedicated context for this simulation.
    ctx, cancel := context.WithCancel(context.Background())
    rt.cancel = cancel

    tickInterval := time.Duration(rt.sim.GetConfig().GetTickRateMs()) * time.Millisecond
    ticker := time.NewTicker(tickInterval)
    defer ticker.Stop()

    // Connect to worker.
    conn, err := grpc.Dial(
        s.workerAddr,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithBlock(),
    )
    if err != nil {
        log.Printf("simulation %s: failed to connect to worker at %s: %v", simID, s.workerAddr, err)
        return
    }
    defer conn.Close()

    workerClient := nodepb.NewNodeWorkerServiceClient(conn)

    stream, err := workerClient.RunWorkerTicks(ctx)
    if err != nil {
        log.Printf("simulation %s: failed to open RunWorkerTicks stream: %v", simID, err)
        return
    }

    log.Printf("simulation %s: tick loop started (entities=%d, tickRateMs=%d)", simID, len(rt.entityIDs), rt.sim.Config.GetTickRateMs())

    var tickNum uint64 = 0

    for {
        select {
        case <-ctx.Done():
            log.Printf("simulation %s: tick loop canceled", simID)
            return
        case <-ticker.C:
            tickNum++

            if err := stream.Send(&nodepb.WorkerTickRequest{
                SimulationId: rt.sim.GetId(),
                Tick:         tickNum,
                PartitionIndex: 0,
                PartitionTotal: 1,
                EntityIds:    rt.entityIDs,
                Config:       rt.sim.GetConfig(),
            }); err != nil {
                log.Printf("simulation %s: send WorkerTickRequest error: %v", simID, err)
                return
            }

            resp, err := stream.Recv()
            if err != nil {
                log.Printf("simulation %s: recv WorkerTickResponse error: %v", simID, err)
                return
            }

            agg := &simulationpb.AggregatedTick{
                SimulationId: rt.sim.GetId(),
                Tick:         resp.GetTick(),
                Entities:     resp.GetEntities(),
                AvgComputeMs: resp.GetComputeMs(),
                WorkerCount:  1, // single worker instance in this v1
                CompletedAt:  timestamppb.New(time.Now()),
            }

            ctxSave, cancelSave := context.WithTimeout(ctx, 2*time.Second)
            _ = s.store.InsertTickSummary(ctxSave, agg)
            cancelSave()

            rt.broadcastTick(agg)
        }
    }
}
