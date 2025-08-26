-- name: GetProject :one
SELECT id, name, created_at, status
FROM projects
WHERE id = $1;

-- name: ListProjects :many
SELECT id, name, created_at, status
FROM projects
ORDER BY created_at DESC;

-- name: CreateProject :one
INSERT INTO projects (name, status)
VALUES ($1, $2)
RETURNING id, name, created_at, status;

-- name: UpdateProjectStatus :one
UPDATE projects 
SET status = $2
WHERE id = $1
RETURNING id, name, created_at, status;
