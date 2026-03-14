-- +goose Up
CREATE TYPE kanban_task_status AS ENUM ('todo', 'progress', 'done', 'failed');

CREATE TABLE kanban_tasks (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title          VARCHAR(255) NOT NULL,
    description    TEXT NOT NULL DEFAULT '',
    agent_task_id  UUID NOT NULL REFERENCES agent_tasks(id) ON DELETE CASCADE,
    status         kanban_task_status NOT NULL DEFAULT 'todo',
    result         TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_kanban_tasks_agent_task_id ON kanban_tasks (agent_task_id);
CREATE INDEX idx_kanban_tasks_status ON kanban_tasks (status);

-- +goose Down
DROP TABLE IF EXISTS kanban_tasks;
DROP TYPE IF EXISTS kanban_task_status;
