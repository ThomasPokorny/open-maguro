-- name: GetTaskExecution :one
SELECT * FROM task_executions WHERE id = $1;

-- name: ListTaskExecutionsByAgentTaskID :many
SELECT * FROM task_executions
WHERE agent_task_id = $1
ORDER BY created_at DESC;

-- name: CreateTaskExecution :one
INSERT INTO task_executions (agent_task_id, status)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateTaskExecutionStatus :one
UPDATE task_executions
SET status = $2,
    started_at = $3,
    finished_at = $4,
    summary = $5,
    error = $6
WHERE id = $1
RETURNING *;
