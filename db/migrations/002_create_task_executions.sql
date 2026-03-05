-- +goose Up
CREATE TYPE execution_status AS ENUM (
    'pending',
    'running',
    'success',
    'failure',
    'timeout'
);

CREATE TABLE task_executions (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_task_id  UUID NOT NULL REFERENCES agent_tasks(id) ON DELETE CASCADE,
    status         execution_status NOT NULL DEFAULT 'pending',
    started_at     TIMESTAMPTZ,
    finished_at    TIMESTAMPTZ,
    summary        TEXT,
    error          TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_task_executions_agent_task_id ON task_executions (agent_task_id);
CREATE INDEX idx_task_executions_status ON task_executions (status);
CREATE INDEX idx_task_executions_created_at ON task_executions (created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS task_executions;
DROP TYPE IF EXISTS execution_status;
