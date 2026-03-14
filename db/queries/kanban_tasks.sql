-- name: CreateKanbanTask :one
INSERT INTO kanban_tasks (title, description, agent_task_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetKanbanTask :one
SELECT * FROM kanban_tasks WHERE id = $1;

-- name: ListKanbanTasks :many
SELECT * FROM kanban_tasks
ORDER BY created_at DESC;

-- name: ListKanbanTasksByAgentID :many
SELECT * FROM kanban_tasks
WHERE agent_task_id = $1
ORDER BY created_at DESC;

-- name: ListKanbanTasksByStatus :many
SELECT * FROM kanban_tasks
WHERE status = $1
ORDER BY created_at DESC;

-- name: ListKanbanTasksByAgentIDAndStatus :many
SELECT * FROM kanban_tasks
WHERE agent_task_id = $1 AND status = $2
ORDER BY created_at DESC;

-- name: ListKanbanTasksByTeamID :many
SELECT kt.* FROM kanban_tasks kt
JOIN agent_tasks at ON kt.agent_task_id = at.id
WHERE at.team_id = $1
ORDER BY kt.created_at DESC;

-- name: ListPendingKanbanTasksByAgentID :many
SELECT * FROM kanban_tasks
WHERE agent_task_id = $1 AND status IN ('todo', 'progress')
ORDER BY created_at ASC;

-- name: UpdateKanbanTask :one
UPDATE kanban_tasks
SET title = $2,
    description = $3,
    agent_task_id = $4,
    status = $5,
    result = $6,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateKanbanTaskStatus :one
UPDATE kanban_tasks
SET status = $2,
    result = $3,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteKanbanTask :exec
DELETE FROM kanban_tasks WHERE id = $1;

-- name: ResetInProgressKanbanTasks :exec
UPDATE kanban_tasks SET status = 'todo', updated_at = now() WHERE status = 'progress';

-- name: ListDistinctAgentsWithPendingKanbanTasks :many
SELECT DISTINCT agent_task_id FROM kanban_tasks
WHERE status IN ('todo', 'progress');
