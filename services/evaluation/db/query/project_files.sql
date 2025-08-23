-- name: GetProjectFile :one
SELECT id, project_id, filename, original_name, file_path, file_size, mime_type, uploaded_at
FROM project_files
WHERE id = $1;

-- name: ListProjectFiles :many
SELECT id, project_id, filename, original_name, file_path, file_size, mime_type, uploaded_at
FROM project_files
WHERE project_id = $1
ORDER BY uploaded_at DESC;

-- name: CreateProjectFile :one
INSERT INTO project_files (project_id, filename, original_name, file_path, file_size, mime_type)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, project_id, filename, original_name, file_path, file_size, mime_type, uploaded_at;

-- name: DeleteProjectFile :exec
DELETE FROM project_files
WHERE id = $1;
