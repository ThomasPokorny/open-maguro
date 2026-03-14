-- name: GetTaskExecution :one
SELECT * FROM task_executions WHERE id = ?;

-- name: ListTaskExecutions :many
SELECT * FROM task_executions
ORDER BY created_at DESC;

-- name: ListTaskExecutionsByAgentTaskID :many
SELECT * FROM task_executions
WHERE agent_task_id = ?
ORDER BY created_at DESC;

-- name: CreateTaskExecution :one
INSERT INTO task_executions (id, agent_task_id, status, task_name, triggered_by_execution_id)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateTaskExecutionStatus :one
UPDATE task_executions
SET status = ?,
    started_at = ?,
    finished_at = ?,
    summary = ?,
    error = ?
WHERE id = ?
RETURNING *;

-- name: GetLatestExecutionByAgentTaskID :one
SELECT * FROM task_executions
WHERE agent_task_id = ?
ORDER BY created_at DESC
LIMIT 1;

-- name: ListStaleRunningExecutions :many
SELECT * FROM task_executions
WHERE status = 'running'
AND started_at < ?;

-- name: DeleteExecutionsOlderThan :execrows
DELETE FROM task_executions
WHERE created_at < ?;
