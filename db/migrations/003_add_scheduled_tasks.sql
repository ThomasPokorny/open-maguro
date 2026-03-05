-- +goose Up
ALTER TABLE agent_tasks
    ADD COLUMN task_type VARCHAR(20) NOT NULL DEFAULT 'cron',
    ADD COLUMN run_at TIMESTAMPTZ,
    ALTER COLUMN cron_expression DROP NOT NULL;

ALTER TABLE task_executions
    ADD COLUMN task_name VARCHAR(255),
    DROP CONSTRAINT task_executions_agent_task_id_fkey,
    ALTER COLUMN agent_task_id DROP NOT NULL,
    ADD CONSTRAINT task_executions_agent_task_id_fkey
        FOREIGN KEY (agent_task_id) REFERENCES agent_tasks(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE task_executions
    DROP CONSTRAINT task_executions_agent_task_id_fkey,
    ALTER COLUMN agent_task_id SET NOT NULL,
    ADD CONSTRAINT task_executions_agent_task_id_fkey
        FOREIGN KEY (agent_task_id) REFERENCES agent_tasks(id) ON DELETE CASCADE,
    DROP COLUMN task_name;

ALTER TABLE agent_tasks
    DROP COLUMN run_at,
    DROP COLUMN task_type,
    ALTER COLUMN cron_expression SET NOT NULL;
