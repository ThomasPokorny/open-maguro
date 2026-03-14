-- name: CreateSkill :one
INSERT INTO skills (id, title, content, environment_secrets)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetSkill :one
SELECT * FROM skills WHERE id = ?;

-- name: ListSkills :many
SELECT * FROM skills ORDER BY created_at DESC;

-- name: UpdateSkill :one
UPDATE skills
SET title = ?,
    content = ?,
    environment_secrets = ?,
    updated_at = datetime('now')
WHERE id = ?
RETURNING *;

-- name: DeleteSkill :exec
DELETE FROM skills WHERE id = ?;

-- name: AddAgentSkill :exec
INSERT INTO agent_skills (agent_task_id, skill_id)
VALUES (?, ?)
ON CONFLICT DO NOTHING;

-- name: RemoveAgentSkill :exec
DELETE FROM agent_skills
WHERE agent_task_id = ? AND skill_id = ?;

-- name: ListSkillsByAgentTaskID :many
SELECT s.* FROM skills s
JOIN agent_skills asj ON s.id = asj.skill_id
WHERE asj.agent_task_id = ?
ORDER BY s.title;
