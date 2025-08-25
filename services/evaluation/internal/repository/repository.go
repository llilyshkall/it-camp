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

// SaveAttach сохраняет информацию о загруженном файле
func (r *Repository) SaveAttach(file *models.Attach) (string, error) {
	// Генерируем уникальное имя файла
	fileName := uuid.New().String() + file.FileExt

	// Создаем запись о файле в базе данных
	// Здесь можно создать таблицу для хранения информации о загруженных файлах
	// Пока возвращаем имя файла для совместимости
	return fileName, nil
}
