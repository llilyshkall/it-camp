package services

import (
	"context"
	"errors"
	db "evaluation/internal/postgres/sqlc"
)

// projectService реализация ProjectService
type projectService struct {
	repo Repository
}

// NewProjectService создает новый экземпляр ProjectService
func NewProjectService(repo Repository) ProjectService {
	return &projectService{
		repo: repo,
	}
}

// CreateProject создает новый проект с валидацией
func (s *projectService) CreateProject(ctx context.Context, name string) (*db.Project, error) {
	// Валидация входных данных
	if name == "" {
		return nil, errors.New("project name is required")
	}

	if len(name) > 255 {
		return nil, errors.New("project name too long (max 255 characters)")
	}

	// Создаем проект через репозиторий
	project, err := s.repo.CreateProject(ctx, name)
	if err != nil {
		return nil, err
	}

	return project, nil
}

// GetProject получает проект по ID
func (s *projectService) GetProject(ctx context.Context, id int32) (*db.Project, error) {
	project, err := s.repo.GetProject(ctx, id)
	if err != nil {
		return nil, err
	}

	return project, nil
}

// ListProjects получает список всех проектов
func (s *projectService) ListProjects(ctx context.Context) ([]db.Project, error) {
	projects, err := s.repo.ListProjects(ctx)
	if err != nil {
		return nil, err
	}

	return projects, nil
}
