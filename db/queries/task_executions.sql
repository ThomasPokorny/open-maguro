-- name: GetTaskExecution :one
SELECT * FROM task_executions WHERE id = $1;

-- name: ListTaskExecutions :many
SELECT * FROM task_executions
ORDER BY created_at DESC;

-- name: ListTaskExecutionsByAgentTaskID :many
SELECT * FROM task_executions
WHERE agent_task_id = $1
ORDER BY created_at DESC;

-- name: CreateTaskExecution :one
INSERT INTO task_executions (agent_task_id, status, task_name, triggered_by_execution_id)
VALUES ($1, $2, $3, $4)
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

-- name: GetLatestExecutionByAgentTaskID :one
SELECT * FROM task_executions
WHERE agent_task_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: ListStaleRunningExecutions :many
SELECT * FROM task_executions
WHERE status = 'running'
AND started_at < $1;

-- name: DeleteExecutionsOlderThan :execrows
DELETE FROM task_executions
WHERE created_at < $1;
