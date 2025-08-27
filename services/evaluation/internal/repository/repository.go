package repository

import (
	"context"
	"evaluation/internal/models"
	"evaluation/internal/postgres"
	db "evaluation/internal/postgres/sqlc"

	"github.com/google/uuid"
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

// CreateProject создает новый проект
func (r *Repository) CreateProject(ctx context.Context, name string) (*db.Project, error) {
	arg := db.CreateProjectParams{
		Name:   name,
		Status: db.ProjectStatusReady, // По умолчанию статус "готов"
	}

	project, err := r.querier.CreateProject(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// ListProjects получает список всех проектов
func (r *Repository) ListProjects(ctx context.Context) ([]db.Project, error) {
	return r.querier.ListProjects(ctx)
}

// GetProject получает проект по ID
func (r *Repository) GetProject(ctx context.Context, id int32) (*db.Project, error) {
	project, err := r.querier.GetProject(ctx, id)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// CheckAndUpdateProjectStatus атомарно проверяет и обновляет статус проекта
// Возвращает ошибку, если проект не найден или статус не "ready"
func (r *Repository) CheckAndUpdateProjectStatus(ctx context.Context, projectID int32, newStatus db.ProjectStatus) (*db.Project, error) {
	arg := db.CheckAndUpdateProjectStatusParams{
		ID:     projectID,
		Status: newStatus,
	}

	project, err := r.querier.CheckAndUpdateProjectStatus(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// UpdateProjectStatus обновляет статус проекта
func (r *Repository) UpdateProjectStatus(ctx context.Context, projectID int32, newStatus db.ProjectStatus) (*db.Project, error) {
	arg := db.UpdateProjectStatusParams{
		ID:     projectID,
		Status: newStatus,
	}

	project, err := r.querier.UpdateProjectStatus(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// GetProjectFilesByType получает файлы проекта по типу
func (r *Repository) GetProjectFilesByType(ctx context.Context, projectID int32, fileType db.FileType) ([]db.ProjectFile, error) {
	arg := db.GetProjectFilesByTypeParams{
		ProjectID: projectID,
		FileType:  fileType,
	}

	files, err := r.querier.GetProjectFilesByType(ctx, arg)
	if err != nil {
		return nil, err
	}
	return files, nil
}

// CreateProjectFile создает запись о файле проекта
func (r *Repository) CreateProjectFile(ctx context.Context, projectID int32, filename, originalName, filePath string, fileSize int64, extension string, fileType db.FileType) (*db.ProjectFile, error) {
	arg := db.CreateProjectFileParams{
		ProjectID:    projectID,
		Filename:     filename,
		OriginalName: originalName,
		FilePath:     filePath,
		FileSize:     fileSize,
		Extension:    extension,
		FileType:     fileType,
	}

	file, err := r.querier.CreateProjectFile(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// CreateRemark создает новое замечание
func (r *Repository) CreateRemark(ctx context.Context, arg db.CreateRemarkParams) (db.Remark, error) {
	return r.querier.CreateRemark(ctx, arg)
}

// SaveAttach сохраняет информацию о загруженном файле
func (r *Repository) SaveAttach(file *models.Attach) (string, error) {
	// Генерируем уникальное имя файла
	fileName := uuid.New().String() + file.FileExt

	// Создаем запись о файле в базе данных
	// Здесь можно создать таблицу для хранения информации о загруженных файлах
	// Пока возвращаем имя файла для совместимости
	return fileName, nil
}
