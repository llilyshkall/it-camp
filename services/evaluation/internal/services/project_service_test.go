package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"evaluation/internal/models"
	db "evaluation/internal/postgres/sqlc"
)

// MockRepository - мок репозитория для тестирования
type MockRepository struct {
	projects map[int32]*db.Project
	nextID   int32
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		projects: make(map[int32]*db.Project),
		nextID:   1,
	}
}

func (m *MockRepository) CreateProject(ctx context.Context, name string) (*db.Project, error) {
	project := &db.Project{
		ID:     m.nextID,
		Name:   name,
		Status: db.ProjectStatusReady,
	}
	m.projects[m.nextID] = project
	m.nextID++
	return project, nil
}

func (m *MockRepository) GetProject(ctx context.Context, id int32) (*db.Project, error) {
	project, exists := m.projects[id]
	if !exists {
		return nil, errors.New("project not found")
	}
	return project, nil
}

func (m *MockRepository) ListProjects(ctx context.Context) ([]db.Project, error) {
	projects := make([]db.Project, 0, len(m.projects))
	for _, project := range m.projects {
		projects = append(projects, *project)
	}
	return projects, nil
}

// Добавляем недостающие методы для реализации интерфейса Repository
func (m *MockRepository) CreateProjectFile(ctx context.Context, projectID int32, filename, originalName, filePath string, fileSize int64, extension string, fileType db.FileType) (*db.ProjectFile, error) {
	// Простая реализация для тестов
	return &db.ProjectFile{
		ID:           1,
		ProjectID:    projectID,
		Filename:     filename,
		OriginalName: originalName,
		FilePath:     filePath,
		FileSize:     fileSize,
		Extension:    extension,
		FileType:     fileType,
	}, nil
}

func (m *MockRepository) SaveAttach(file *models.Attach) (string, error) {
	// Простая реализация для тестов
	return "mock-filename" + file.FileExt, nil
}

func (m *MockRepository) CheckAndUpdateProjectStatus(ctx context.Context, projectID int32, newStatus db.ProjectStatus) (*db.Project, error) {
	project, exists := m.projects[projectID]
	if !exists {
		return nil, errors.New("project not found")
	}
	project.Status = newStatus
	return project, nil
}

func (m *MockRepository) GetProjectFilesByType(ctx context.Context, projectID int32, fileType db.FileType) ([]db.ProjectFile, error) {
	// Простая реализация для тестов - возвращаем пустой список
	return []db.ProjectFile{}, nil
}

func (m *MockRepository) UpdateProjectStatus(ctx context.Context, projectID int32, newStatus db.ProjectStatus) (*db.Project, error) {
	project, exists := m.projects[projectID]
	if !exists {
		return nil, errors.New("project not found")
	}
	project.Status = newStatus
	return project, nil
}

func (m *MockRepository) CreateRemark(ctx context.Context, arg db.CreateRemarkParams) (db.Remark, error) {
	// Простая реализация для тестов
	return db.Remark{
		ID:         1,
		ProjectID:  arg.ProjectID,
		Direction:  arg.Direction,
		Section:    arg.Section,
		Subsection: arg.Subsection,
		Content:    arg.Content,
		CreatedAt:  time.Now(),
	}, nil
}

// Тесты для ProjectService
func TestProjectService_CreateProject(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid project name",
			projectName: "Test Project",
			wantErr:     false,
		},
		{
			name:        "empty project name",
			projectName: "",
			wantErr:     true,
			errContains: "project name is required",
		},
		{
			name:        "project name too long",
			projectName: string(make([]byte, 256)), // 256 символов
			wantErr:     true,
			errContains: "project name too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockRepository()
			service := NewProjectService(mockRepo)

			project, err := service.CreateProject(context.Background(), tt.projectName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.errContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if project == nil {
					t.Errorf("Expected project but got nil")
					return
				}
				if project.Name != tt.projectName {
					t.Errorf("Expected project name '%s', got '%s'", tt.projectName, project.Name)
				}
				if project.Status != db.ProjectStatusReady {
					t.Errorf("Expected project status '%s', got '%s'", db.ProjectStatusReady, project.Status)
				}
			}
		})
	}
}

func TestProjectService_GetProject(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewProjectService(mockRepo)

	// Создаем тестовый проект
	createdProject, err := service.CreateProject(context.Background(), "Test Project")
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// Тестируем получение существующего проекта
	project, err := service.GetProject(context.Background(), createdProject.ID)
	if err != nil {
		t.Errorf("Failed to get existing project: %v", err)
	}
	if project.ID != createdProject.ID {
		t.Errorf("Expected project ID %d, got %d", createdProject.ID, project.ID)
	}

	// Тестируем получение несуществующего проекта
	_, err = service.GetProject(context.Background(), 999)
	if err == nil {
		t.Errorf("Expected error when getting non-existent project")
	}
}

func TestProjectService_ListProjects(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewProjectService(mockRepo)

	// Создаем несколько тестовых проектов
	projectNames := []string{"Project 1", "Project 2", "Project 3"}
	for _, name := range projectNames {
		_, err := service.CreateProject(context.Background(), name)
		if err != nil {
			t.Fatalf("Failed to create test project '%s': %v", name, err)
		}
	}

	// Тестируем получение списка проектов
	projects, err := service.ListProjects(context.Background())
	if err != nil {
		t.Errorf("Failed to list projects: %v", err)
	}

	if len(projects) != len(projectNames) {
		t.Errorf("Expected %d projects, got %d", len(projectNames), len(projects))
	}

	// Проверяем, что все проекты присутствуют
	projectMap := make(map[string]bool)
	for _, project := range projects {
		projectMap[project.Name] = true
	}

	for _, name := range projectNames {
		if !projectMap[name] {
			t.Errorf("Project '%s' not found in list", name)
		}
	}
}
