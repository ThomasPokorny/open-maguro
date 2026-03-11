-- +goose Up
ALTER TABLE agent_tasks ADD COLUMN on_success_task_id UUID REFERENCES agent_tasks(id) ON DELETE SET NULL;
ALTER TABLE agent_tasks ADD COLUMN on_failure_task_id UUID REFERENCES agent_tasks(id) ON DELETE SET NULL;
ALTER TABLE task_executions ADD COLUMN triggered_by_execution_id UUID REFERENCES task_executions(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE task_executions DROP COLUMN triggered_by_execution_id;
ALTER TABLE agent_tasks DROP COLUMN on_failure_task_id;
ALTER TABLE agent_tasks DROP COLUMN on_success_task_id;
