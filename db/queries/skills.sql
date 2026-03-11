-- name: CreateSkill :one
INSERT INTO skills (title, content)
VALUES ($1, $2)
RETURNING *;

-- name: GetSkill :one
SELECT * FROM skills WHERE id = $1;

-- name: ListSkills :many
SELECT * FROM skills ORDER BY created_at DESC;

-- name: UpdateSkill :one
UPDATE skills
SET title = $2,
    content = $3,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteSkill :exec
DELETE FROM skills WHERE id = $1;

-- name: AddAgentSkill :exec
INSERT INTO agent_skills (agent_task_id, skill_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveAgentSkill :exec
DELETE FROM agent_skills
WHERE agent_task_id = $1 AND skill_id = $2;

-- name: ListSkillsByAgentTaskID :many
SELECT s.* FROM skills s
JOIN agent_skills asj ON s.id = asj.skill_id
WHERE asj.agent_task_id = $1
ORDER BY s.title;
