-- +goose Up
CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    color VARCHAR(7) NOT NULL DEFAULT '#6366f1',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE agent_tasks ADD COLUMN team_id UUID REFERENCES teams(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE agent_tasks DROP COLUMN team_id;
DROP TABLE teams;
