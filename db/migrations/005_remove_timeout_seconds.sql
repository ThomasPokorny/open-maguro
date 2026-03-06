-- +goose Up
ALTER TABLE agent_tasks DROP COLUMN timeout_seconds;

-- +goose Down
ALTER TABLE agent_tasks ADD COLUMN timeout_seconds INTEGER NOT NULL DEFAULT 60;
