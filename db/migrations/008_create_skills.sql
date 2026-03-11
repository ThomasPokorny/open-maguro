-- +goose Up
CREATE TABLE skills (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title      VARCHAR(255) NOT NULL,
    content    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE agent_skills (
    agent_task_id UUID NOT NULL REFERENCES agent_tasks(id) ON DELETE CASCADE,
    skill_id      UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    PRIMARY KEY (agent_task_id, skill_id)
);

ALTER TABLE agent_tasks ADD COLUMN global_skill_access BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE agent_tasks DROP COLUMN global_skill_access;
DROP TABLE IF EXISTS agent_skills;
DROP TABLE IF EXISTS skills;
