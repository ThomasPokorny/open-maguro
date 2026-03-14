-- name: CreateKanbanTask :one
INSERT INTO kanban_tasks (id, title, description, agent_task_id)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetKanbanTask :one
SELECT * FROM kanban_tasks WHERE id = ?;

-- name: ListKanbanTasks :many
SELECT * FROM kanban_tasks
ORDER BY created_at DESC;

-- name: ListKanbanTasksByAgentID :many
SELECT * FROM kanban_tasks
WHERE agent_task_id = ?
ORDER BY created_at DESC;

-- name: ListKanbanTasksByStatus :many
SELECT * FROM kanban_tasks
WHERE status = ?
ORDER BY created_at DESC;

-- name: ListKanbanTasksByAgentIDAndStatus :many
SELECT * FROM kanban_tasks
WHERE agent_task_id = ? AND status = ?
ORDER BY created_at DESC;

-- name: ListKanbanTasksByTeamID :many
SELECT kt.* FROM kanban_tasks kt
JOIN agent_tasks at ON kt.agent_task_id = at.id
WHERE at.team_id = ?
ORDER BY kt.created_at DESC;

-- name: ListPendingKanbanTasksByAgentID :many
SELECT * FROM kanban_tasks
WHERE agent_task_id = ? AND status IN ('todo', 'progress')
ORDER BY created_at ASC;

-- name: UpdateKanbanTask :one
UPDATE kanban_tasks
SET title = ?,
    description = ?,
    agent_task_id = ?,
    status = ?,
    result = ?,
    updated_at = datetime('now')
WHERE id = ?
RETURNING *;

-- name: UpdateKanbanTaskStatus :one
UPDATE kanban_tasks
SET status = ?,
    result = ?,
    updated_at = datetime('now')
WHERE id = ?
RETURNING *;

-- name: DeleteKanbanTask :exec
DELETE FROM kanban_tasks WHERE id = ?;

-- name: ResetInProgressKanbanTasks :exec
UPDATE kanban_tasks SET status = 'todo', updated_at = datetime('now') WHERE status = 'progress';

-- name: ListDistinctAgentsWithPendingKanbanTasks :many
SELECT DISTINCT agent_task_id FROM kanban_tasks
WHERE status IN ('todo', 'progress');
