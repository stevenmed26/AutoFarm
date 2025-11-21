// internal/store/simulation_store.go
package store

import (
	"context"
	"database/sql"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	commonpb "github.com/stevenmed26/AutoFarm/internal/proto/commonpb"
	simulationpb "github.com/stevenmed26/AutoFarm/internal/proto/simulationpb"
)

type SimulationStore struct {
	db *DB
}

func NewSimulationStore(db *DB) *SimulationStore {
	return &SimulationStore{db: db}
}

func (s *SimulationStore) InsertSimulation(ctx context.Context, sim *simulationpb.Simulation) error {
	_, err := s.db.sql.ExecContext(ctx, `
		INSERT INTO simulations (
			id, name, entity_count, tick_rate_ms, scenario_type,
			status, created_at, started_at, ended_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`,
		sim.GetId().GetValue(),
		sim.GetConfig().GetName(),
		sim.GetConfig().GetEntityCount(),
		sim.GetConfig().GetTickRateMs(),
		sim.GetConfig().GetScenarioType(),
		sim.GetStatus().String(),
		toTime(sim.GetCreatedAt()),
		toTimePtr(sim.GetStartedAt()),
		toTimePtr(sim.GetEndedAt()),
	)
	return err
}

func (s *SimulationStore) UpdateSimulationStatus(ctx context.Context, sim *simulationpb.Simulation) error {
	_, err := s.db.sql.ExecContext(ctx, `
		UPDATE simulations
		SET status = $2,
		    started_at = $3,
		    ended_at = $4
		WHERE id = $1
	`,
		sim.GetId().GetValue(),
		sim.GetStatus().String(),
		toTimePtr(sim.GetStartedAt()),
		toTimePtr(sim.GetEndedAt()),
	)
	return err
}

func (s *SimulationStore) GetSimulation(ctx context.Context, id *commonpb.SimulationId) (*simulationpb.Simulation, error) {
	row := s.db.sql.QueryRowContext(ctx, `
		SELECT
			id, name, entity_count, tick_rate_ms, scenario_type,
			status, created_at, started_at, ended_at
		FROM simulations
		WHERE id = $1
	`, id.GetValue())

	var simID, name, scenarioType, statusText string
	var entityCount, tickRate int32
	var createdAt time.Time
	var startedAt, endedAt sql.NullTime

	if err := row.Scan(
		&simID, &name, &entityCount, &tickRate, &scenarioType,
		&statusText, &createdAt, &startedAt, &endedAt,
	); err != nil {
		return nil, err
	}

	status := commonpb.SimulationStatus_SIMULATION_STATUS_UNSPECIFIED
	for k, v := range commonpb.SimulationStatus_name {
		if v == statusText {
			status = commonpb.SimulationStatus(k)
			break
		}
	}

	return &simulationpb.Simulation{
		Id: &commonpb.SimulationId{Value: simID},
		Config: &simulationpb.SimulationConfig{
			Name:         name,
			EntityCount:  uint32(entityCount),
			TickRateMs:   uint32(tickRate),
			ScenarioType: scenarioType,
		},
		Status:    status,
		CreatedAt: toTimestamp(createdAt),
		StartedAt: toTimestampNull(startedAt),
		EndedAt:   toTimestampNull(endedAt),
	}, nil
}

func (s *SimulationStore) InsertTickSummary(ctx context.Context, tick *simulationpb.AggregatedTick) error {
	_, err := s.db.sql.ExecContext(ctx, `
		INSERT INTO simulation_ticks (
			simulation_id, tick, entity_count,
			avg_compute_ms, worker_count, completed_at
		)
		VALUES ($1,$2,$3,$4,$5,$6)
	`,
		tick.GetSimulationId().GetValue(),
		tick.GetTick(),
		len(tick.GetEntities()),
		tick.GetAvgComputeMs(),
		tick.GetWorkerCount(),
		toTime(tick.GetCompletedAt()),
	)
	return err
}

// helpers

func toTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}

func toTimePtr(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

func toTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func toTimestampNull(nt sql.NullTime) *timestamppb.Timestamp {
	if !nt.Valid {
		return nil
	}
	return timestamppb.New(nt.Time)
}
