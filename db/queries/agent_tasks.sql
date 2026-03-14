-- name: GetAgentTask :one
SELECT * FROM agent_tasks WHERE id = ?;

-- name: ListAgentTasks :many
SELECT * FROM agent_tasks ORDER BY created_at DESC;

-- name: ListUserAgentTasks :many
SELECT * FROM agent_tasks
WHERE system_agent = false
ORDER BY created_at DESC;

-- name: ListSystemAgentTasks :many
SELECT * FROM agent_tasks
WHERE system_agent = true
ORDER BY created_at DESC;

-- name: ListEnabledCronTasks :many
SELECT * FROM agent_tasks
WHERE enabled = true AND task_type = 'cron'
ORDER BY created_at DESC;

-- name: ListPendingScheduledTasks :many
SELECT * FROM agent_tasks
WHERE enabled = true AND task_type = 'one_time'
ORDER BY run_at ASC;

-- name: ListAgentTasksByTeamID :many
SELECT * FROM agent_tasks
WHERE team_id = ?
ORDER BY created_at DESC;

-- name: CreateAgentTask :one
INSERT INTO agent_tasks (id, name, cron_expression, prompt, enabled, mcp_config, allowed_tools, system_agent, global_skill_access, on_success_task_id, on_failure_task_id, team_id, task_type)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'cron')
RETURNING *;

-- name: CreateScheduledTask :one
INSERT INTO agent_tasks (id, name, prompt, run_at, mcp_config, allowed_tools, system_agent, global_skill_access, task_type)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'one_time')
RETURNING *;

-- name: UpdateAgentTask :one
UPDATE agent_tasks
SET name = ?,
    cron_expression = ?,
    prompt = ?,
    enabled = ?,
    mcp_config = ?,
    allowed_tools = ?,
    system_agent = ?,
    global_skill_access = ?,
    on_success_task_id = ?,
    on_failure_task_id = ?,
    team_id = ?,
    updated_at = datetime('now')
WHERE id = ?
RETURNING *;

-- name: DeleteAgentTask :exec
DELETE FROM agent_tasks WHERE id = ?;
