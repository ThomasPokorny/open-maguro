-- name: GetTeam :one
SELECT * FROM teams WHERE id = $1;

-- name: ListTeams :many
SELECT * FROM teams ORDER BY created_at DESC;

-- name: CreateTeam :one
INSERT INTO teams (title, description, color)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateTeam :one
UPDATE teams
SET title = $2,
    description = $3,
    color = $4,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteTeam :exec
DELETE FROM teams WHERE id = $1;
