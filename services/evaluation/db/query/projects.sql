-- name: GetProject :one
SELECT id, name, created_at, in_progress
FROM projects
WHERE id = $1;

-- name: ListProjects :many
SELECT id, name, created_at, in_progress
FROM projects
ORDER BY created_at DESC;

-- name: CreateProject :one
INSERT INTO projects (name, in_progress)
VALUES ($1, $2)
RETURNING id, name, created_at, in_progress;

-- name: UpdateProject :one
UPDATE projects
SET name = $2, in_progress = $3
WHERE id = $1
RETURNING id, name, created_at, in_progress;

-- name: DeleteProject :exec
DELETE FROM projects
WHERE id = $1;
