-- name: GetAgentTask :one
SELECT * FROM agent_tasks WHERE id = $1;

-- name: ListAgentTasks :many
SELECT * FROM agent_tasks ORDER BY created_at DESC;

-- name: ListEnabledAgentTasks :many
SELECT * FROM agent_tasks WHERE enabled = true ORDER BY created_at DESC;

-- name: CreateAgentTask :one
INSERT INTO agent_tasks (name, cron_expression, prompt, enabled, timeout_seconds)
VALUES ($1, $2, $3, $4, $5)
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
