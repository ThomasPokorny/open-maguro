-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE agent_tasks (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name             VARCHAR(255) NOT NULL,
    cron_expression  VARCHAR(100) NOT NULL,
    prompt           TEXT NOT NULL,
    enabled          BOOLEAN NOT NULL DEFAULT true,
    timeout_seconds  INTEGER NOT NULL DEFAULT 60,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_agent_tasks_enabled ON agent_tasks (enabled);

-- +goose Down
DROP TABLE IF EXISTS agent_tasks;
