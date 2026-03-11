-- +goose Up
ALTER TABLE agent_tasks ADD COLUMN system_agent BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE agent_tasks DROP COLUMN system_agent;
