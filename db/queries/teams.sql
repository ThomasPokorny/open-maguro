-- name: GetTeam :one
SELECT * FROM teams WHERE id = ?;

-- name: ListTeams :many
SELECT * FROM teams ORDER BY created_at DESC;

-- name: CreateTeam :one
INSERT INTO teams (id, title, description, color)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: UpdateTeam :one
UPDATE teams
SET title = ?,
    description = ?,
    color = ?,
    updated_at = datetime('now')
WHERE id = ?
RETURNING *;

-- name: DeleteTeam :exec
DELETE FROM teams WHERE id = ?;
