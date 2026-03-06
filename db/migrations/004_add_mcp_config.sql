-- +goose Up
ALTER TABLE agent_tasks ADD COLUMN mcp_config TEXT;

-- +goose Down
ALTER TABLE agent_tasks DROP COLUMN mcp_config;
