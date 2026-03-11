-- +goose Up
ALTER TABLE agent_tasks ADD COLUMN allowed_tools TEXT;

-- +goose Down
ALTER TABLE agent_tasks DROP COLUMN allowed_tools;
