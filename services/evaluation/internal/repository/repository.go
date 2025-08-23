package repository

import (
	"context"
	"database/sql"
	"evaluation/internal/postgres"
	db "evaluation/internal/postgres/sqlc"
)

// Repository объединяет все операции с базой данных
type Repository struct {
	querier db.Querier
}

// New создает новый экземпляр репозитория
func New(pgClient *postgres.Client) *Repository {
	querier := db.New(pgClient.DB)
	return &Repository{querier: querier}
}

// ========== PROJECT METHODS ==========

// GetProject получает проект по ID
func (r *Repository) GetProject(ctx context.Context, id int32) (*db.Project, error) {
	project, err := r.querier.GetProject(ctx, id)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// ListProjects получает список всех проектов
func (r *Repository) ListProjects(ctx context.Context) ([]db.Project, error) {
	return r.querier.ListProjects(ctx)
}

// CreateProject создает новый проект
func (r *Repository) CreateProject(ctx context.Context, name string, inProgress bool) (*db.Project, error) {
	arg := db.CreateProjectParams{
		Name:       name,
		InProgress: inProgress,
	}
	
	project, err := r.querier.CreateProject(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// UpdateProject обновляет проект
func (r *Repository) UpdateProject(ctx context.Context, id int32, name string, inProgress bool) (*db.Project, error) {
	arg := db.UpdateProjectParams{
		ID:         id,
		Name:       name,
		InProgress: inProgress,
	}
	
	project, err := r.querier.UpdateProject(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// DeleteProject удаляет проект
func (r *Repository) DeleteProject(ctx context.Context, id int32) error {
	return r.querier.DeleteProject(ctx, id)
}

// ========== REMARK METHODS ==========

// GetRemark получает замечание по ID
func (r *Repository) GetRemark(ctx context.Context, id int32) (*db.Remark, error) {
	remark, err := r.querier.GetRemark(ctx, id)
	if err != nil {
		return nil, err
	}
	return &remark, nil
}

// ListRemarksByProject получает список замечаний по проекту
func (r *Repository) ListRemarksByProject(ctx context.Context, projectID int32) ([]db.Remark, error) {
	return r.querier.ListRemarksByProject(ctx, projectID)
}

// CreateRemark создает новое замечание
func (r *Repository) CreateRemark(ctx context.Context, projectID int32, direction, section, subsection, content string) (*db.Remark, error) {
	arg := db.CreateRemarkParams{
		ProjectID:  projectID,
		Direction:  direction,
		Section:    section,
		Subsection: subsection,
		Content:    content,
	}
	
	remark, err := r.querier.CreateRemark(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &remark, nil
}

// UpdateRemark обновляет замечание
func (r *Repository) UpdateRemark(ctx context.Context, id int32, direction, section, subsection, content string) (*db.Remark, error) {
	arg := db.UpdateRemarkParams{
		ID:         id,
		Direction:  direction,
		Section:    section,
		Subsection: subsection,
		Content:    content,
	}
	
	remark, err := r.querier.UpdateRemark(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &remark, nil
}

// DeleteRemark удаляет замечание
func (r *Repository) DeleteRemark(ctx context.Context, id int32) error {
	return r.querier.DeleteRemark(ctx, id)
}

// ========== PROJECT FILE METHODS ==========

// GetProjectFile получает файл проекта по ID
func (r *Repository) GetProjectFile(ctx context.Context, id int32) (*db.ProjectFile, error) {
	file, err := r.querier.GetProjectFile(ctx, id)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// ListProjectFiles получает список файлов проекта
func (r *Repository) ListProjectFiles(ctx context.Context, projectID int32) ([]db.ProjectFile, error) {
	return r.querier.ListProjectFiles(ctx, projectID)
}

// CreateProjectFile создает новый файл проекта
func (r *Repository) CreateProjectFile(ctx context.Context, projectID int32, filename, originalName, filePath string, fileSize int64, mimeType string) (*db.ProjectFile, error) {
	arg := db.CreateProjectFileParams{
		ProjectID:    projectID,
		Filename:     filename,
		OriginalName: originalName,
		FilePath:     filePath,
		FileSize:     fileSize,
		MimeType:     sql.NullString{String: mimeType, Valid: mimeType != ""},
	}
	
	file, err := r.querier.CreateProjectFile(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// DeleteProjectFile удаляет файл проекта
func (r *Repository) DeleteProjectFile(ctx context.Context, id int32) error {
	return r.querier.DeleteProjectFile(ctx, id)
}
