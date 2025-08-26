-- name: CreateProjectFile :one
INSERT INTO project_files (project_id, filename, original_name, file_path, file_size, extension, file_type)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, project_id, filename, original_name, file_path, file_size, extension, file_type, uploaded_at;

-- name: GetProjectFiles :many
SELECT id, project_id, filename, original_name, file_path, file_size, extension, file_type, uploaded_at
FROM project_files
WHERE project_id = $1
ORDER BY uploaded_at DESC;

-- name: GetProjectFilesByType :many
SELECT id, project_id, filename, original_name, file_path, file_size, extension, file_type, uploaded_at
FROM project_files
WHERE project_id = $1 AND file_type = $2
ORDER BY uploaded_at DESC;
