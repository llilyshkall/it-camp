-- name: CreateProjectFile :one
INSERT INTO project_files (project_id, filename, original_name, file_path, file_size, extension, file_type)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, project_id, filename, original_name, file_path, file_size, extension, file_type, uploaded_at;
