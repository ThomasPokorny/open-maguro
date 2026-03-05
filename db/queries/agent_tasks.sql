-- name: GetAgentTask :one
SELECT * FROM agent_tasks WHERE id = $1;

-- name: ListAgentTasks :many
SELECT * FROM agent_tasks ORDER BY created_at DESC;

-- name: ListEnabledCronTasks :many
SELECT * FROM agent_tasks
WHERE enabled = true AND task_type = 'cron'
ORDER BY created_at DESC;

-- name: ListPendingScheduledTasks :many
SELECT * FROM agent_tasks
WHERE enabled = true AND task_type = 'one_time'
ORDER BY run_at ASC;

-- name: CreateAgentTask :one
INSERT INTO agent_tasks (name, cron_expression, prompt, enabled, timeout_seconds, task_type)
VALUES ($1, $2, $3, $4, $5, 'cron')
RETURNING *;

-- name: CreateScheduledTask :one
INSERT INTO agent_tasks (name, prompt, run_at, timeout_seconds, task_type)
VALUES ($1, $2, $3, $4, 'one_time')
RETURNING *;

-- name: UpdateAgentTask :one
UPDATE agent_tasks
SET name = $2,
    cron_expression = $3,
    prompt = $4,
    enabled = $5,
    timeout_seconds = $6,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteAgentTask :exec
DELETE FROM agent_tasks WHERE id = $1;
