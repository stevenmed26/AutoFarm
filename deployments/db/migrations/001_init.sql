-- deployments/db/migrations/001_init.sql

CREATE TABLE IF NOT EXISTS simulations (
    id             uuid PRIMARY KEY,
    name           text NOT NULL,
    entity_count   integer NOT NULL,
    tick_rate_ms   integer NOT NULL,
    scenario_type  text NOT NULL,
    status         text NOT NULL,
    created_at     timestamptz NOT NULL,
    started_at     timestamptz,
    ended_at       timestamptz
);

CREATE INDEX IF NOT EXISTS idx_simulations_status ON simulations (status);

CREATE TABLE IF NOT EXISTS simulation_ticks (
    id             bigserial PRIMARY KEY,
    simulation_id  uuid NOT NULL REFERENCES simulations(id) ON DELETE CASCADE,
    tick           bigint NOT NULL,
    entity_count   integer NOT NULL,
    avg_compute_ms double precision NOT NULL,
    worker_count   integer NOT NULL,
    completed_at   timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_simulation_ticks_simulation_id ON simulation_ticks (simulation_id);
CREATE INDEX IF NOT EXISTS idx_simulation_ticks_tick ON simulation_ticks (simulation_id, tick);
