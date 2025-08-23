-- name: GetRemark :one
SELECT id, project_id, direction, section, subsection, content, created_at
FROM remarks
WHERE id = $1;

-- name: ListRemarksByProject :many
SELECT id, project_id, direction, section, subsection, content, created_at
FROM remarks
WHERE project_id = $1
ORDER BY created_at DESC;

-- name: CreateRemark :one
INSERT INTO remarks (project_id, direction, section, subsection, content)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, project_id, direction, section, subsection, content, created_at;

-- name: UpdateRemark :one
UPDATE remarks
SET direction = $2, section = $3, subsection = $4, content = $5
WHERE id = $1
RETURNING id, project_id, direction, section, subsection, content, created_at;

-- name: DeleteRemark :exec
DELETE FROM remarks
WHERE id = $1;
