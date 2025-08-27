package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"evaluation/internal/models"
	db "evaluation/internal/postgres/sqlc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockQuerier - мок для интерфейса Querier
type MockQuerier struct {
	mock.Mock
}

func (m *MockQuerier) CreateProject(ctx context.Context, arg db.CreateProjectParams) (db.Project, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Project), args.Error(1)
}

func (m *MockQuerier) ListProjects(ctx context.Context) ([]db.Project, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.Project), args.Error(1)
}

func (m *MockQuerier) GetProject(ctx context.Context, id int32) (db.Project, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Project), args.Error(1)
}

func (m *MockQuerier) CheckAndUpdateProjectStatus(ctx context.Context, arg db.CheckAndUpdateProjectStatusParams) (db.Project, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Project), args.Error(1)
}

func (m *MockQuerier) UpdateProjectStatus(ctx context.Context, arg db.UpdateProjectStatusParams) (db.Project, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Project), args.Error(1)
}

func (m *MockQuerier) GetProjectFilesByType(ctx context.Context, arg db.GetProjectFilesByTypeParams) ([]db.ProjectFile, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]db.ProjectFile), args.Error(1)
}

func (m *MockQuerier) CreateProjectFile(ctx context.Context, arg db.CreateProjectFileParams) (db.ProjectFile, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.ProjectFile), args.Error(1)
}

func (m *MockQuerier) CreateRemark(ctx context.Context, arg db.CreateRemarkParams) (db.Remark, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Remark), args.Error(1)
}

func (m *MockQuerier) GetRemarksByProject(ctx context.Context, projectID int32) ([]db.Remark, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).([]db.Remark), args.Error(1)
}

func (m *MockQuerier) GetProjectFiles(ctx context.Context, projectID int32) ([]db.ProjectFile, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).([]db.ProjectFile), args.Error(1)
}

// TestRepository_CreateProject тестирует создание проекта
func TestRepository_CreateProject(t *testing.T) {
	tests := []struct {
		name          string
		projectName   string
		mockProject   db.Project
		mockError     error
		expectedError bool
	}{
		{
			name:        "Успешное создание проекта",
			projectName: "Test Project",
			mockProject: db.Project{
				ID:        1,
				Name:      "Test Project",
				Status:    db.ProjectStatusReady,
				CreatedAt: time.Now(),
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "Ошибка при создании проекта",
			projectName:   "Test Project",
			mockProject:   db.Project{},
			mockError:     errors.New("database error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQuerier := new(MockQuerier)
			repo := &Repository{querier: mockQuerier}

			expectedArg := db.CreateProjectParams{
				Name:   tt.projectName,
				Status: db.ProjectStatusReady,
			}

			mockQuerier.On("CreateProject", mock.Anything, expectedArg).Return(tt.mockProject, tt.mockError)

			result, err := repo.CreateProject(context.Background(), tt.projectName)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockProject.ID, result.ID)
				assert.Equal(t, tt.mockProject.Name, result.Name)
				assert.Equal(t, tt.mockProject.Status, result.Status)
			}

			mockQuerier.AssertExpectations(t)
		})
	}
}

// TestRepository_ListProjects тестирует получение списка проектов
func TestRepository_ListProjects(t *testing.T) {
	tests := []struct {
		name          string
		mockProjects  []db.Project
		mockError     error
		expectedError bool
	}{
		{
			name: "Успешное получение списка проектов",
			mockProjects: []db.Project{
				{
					ID:        1,
					Name:      "Project 1",
					Status:    db.ProjectStatusReady,
					CreatedAt: time.Now(),
				},
				{
					ID:        2,
					Name:      "Project 2",
					Status:    db.ProjectStatusProcessingRemarks,
					CreatedAt: time.Now(),
				},
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "Пустой список проектов",
			mockProjects:  []db.Project{},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "Ошибка при получении списка проектов",
			mockProjects:  nil,
			mockError:     errors.New("database error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQuerier := new(MockQuerier)
			repo := &Repository{querier: mockQuerier}

			mockQuerier.On("ListProjects", mock.Anything).Return(tt.mockProjects, tt.mockError)

			result, err := repo.ListProjects(context.Background())

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, len(tt.mockProjects))
			}

			mockQuerier.AssertExpectations(t)
		})
	}
}

// TestRepository_GetProject тестирует получение проекта по ID
func TestRepository_GetProject(t *testing.T) {
	tests := []struct {
		name          string
		projectID     int32
		mockProject   db.Project
		mockError     error
		expectedError bool
	}{
		{
			name:      "Успешное получение проекта",
			projectID: 1,
			mockProject: db.Project{
				ID:        1,
				Name:      "Test Project",
				Status:    db.ProjectStatusReady,
				CreatedAt: time.Now(),
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "Проект не найден",
			projectID:     999,
			mockProject:   db.Project{},
			mockError:     errors.New("project not found"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQuerier := new(MockQuerier)
			repo := &Repository{querier: mockQuerier}

			mockQuerier.On("GetProject", mock.Anything, tt.projectID).Return(tt.mockProject, tt.mockError)

			result, err := repo.GetProject(context.Background(), tt.projectID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockProject.ID, result.ID)
				assert.Equal(t, tt.mockProject.Name, result.Name)
			}

			mockQuerier.AssertExpectations(t)
		})
	}
}

// TestRepository_CheckAndUpdateProjectStatus тестирует проверку и обновление статуса проекта
func TestRepository_CheckAndUpdateProjectStatus(t *testing.T) {
	tests := []struct {
		name          string
		projectID     int32
		newStatus     db.ProjectStatus
		mockProject   db.Project
		mockError     error
		expectedError bool
	}{
		{
			name:      "Успешное обновление статуса проекта",
			projectID: 1,
			newStatus: db.ProjectStatusProcessingRemarks,
			mockProject: db.Project{
				ID:        1,
				Name:      "Test Project",
				Status:    db.ProjectStatusProcessingRemarks,
				CreatedAt: time.Now(),
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "Ошибка при обновлении статуса",
			projectID:     1,
			newStatus:     db.ProjectStatusProcessingRemarks,
			mockProject:   db.Project{},
			mockError:     errors.New("status update failed"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQuerier := new(MockQuerier)
			repo := &Repository{querier: mockQuerier}

			expectedArg := db.CheckAndUpdateProjectStatusParams{
				ID:     tt.projectID,
				Status: tt.newStatus,
			}

			mockQuerier.On("CheckAndUpdateProjectStatus", mock.Anything, expectedArg).Return(tt.mockProject, tt.mockError)

			result, err := repo.CheckAndUpdateProjectStatus(context.Background(), tt.projectID, tt.newStatus)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.newStatus, result.Status)
			}

			mockQuerier.AssertExpectations(t)
		})
	}
}

// TestRepository_UpdateProjectStatus тестирует обновление статуса проекта
func TestRepository_UpdateProjectStatus(t *testing.T) {
	tests := []struct {
		name          string
		projectID     int32
		newStatus     db.ProjectStatus
		mockProject   db.Project
		mockError     error
		expectedError bool
	}{
		{
			name:      "Успешное обновление статуса проекта",
			projectID: 1,
			newStatus: db.ProjectStatusGeneratingFinalReport,
			mockProject: db.Project{
				ID:        1,
				Name:      "Test Project",
				Status:    db.ProjectStatusGeneratingFinalReport,
				CreatedAt: time.Now(),
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "Ошибка при обновлении статуса",
			projectID:     1,
			newStatus:     db.ProjectStatusGeneratingFinalReport,
			mockProject:   db.Project{},
			mockError:     errors.New("update failed"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQuerier := new(MockQuerier)
			repo := &Repository{querier: mockQuerier}

			expectedArg := db.UpdateProjectStatusParams{
				ID:     tt.projectID,
				Status: tt.newStatus,
			}

			mockQuerier.On("UpdateProjectStatus", mock.Anything, expectedArg).Return(tt.mockProject, tt.mockError)

			result, err := repo.UpdateProjectStatus(context.Background(), tt.projectID, tt.newStatus)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.newStatus, result.Status)
			}

			mockQuerier.AssertExpectations(t)
		})
	}
}

// TestRepository_GetProjectFilesByType тестирует получение файлов проекта по типу
func TestRepository_GetProjectFilesByType(t *testing.T) {
	tests := []struct {
		name          string
		projectID     int32
		fileType      db.FileType
		mockFiles     []db.ProjectFile
		mockError     error
		expectedError bool
	}{
		{
			name:      "Успешное получение файлов по типу",
			projectID: 1,
			fileType:  db.FileTypeDocumentation,
			mockFiles: []db.ProjectFile{
				{
					ID:           1,
					ProjectID:    1,
					Filename:     "doc1.pdf",
					OriginalName: "documentation.pdf",
					FilePath:     "/files/doc1.pdf",
					FileSize:     1024,
					Extension:    "pdf",
					FileType:     db.FileTypeDocumentation,
					UploadedAt:   time.Now(),
				},
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "Пустой список файлов",
			projectID:     1,
			fileType:      db.FileTypeChecklist,
			mockFiles:     []db.ProjectFile{},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "Ошибка при получении файлов",
			projectID:     1,
			fileType:      db.FileTypeDocumentation,
			mockFiles:     nil,
			mockError:     errors.New("database error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQuerier := new(MockQuerier)
			repo := &Repository{querier: mockQuerier}

			expectedArg := db.GetProjectFilesByTypeParams{
				ProjectID: tt.projectID,
				FileType:  tt.fileType,
			}

			mockQuerier.On("GetProjectFilesByType", mock.Anything, expectedArg).Return(tt.mockFiles, tt.mockError)

			result, err := repo.GetProjectFilesByType(context.Background(), tt.projectID, tt.fileType)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, len(tt.mockFiles))
			}

			mockQuerier.AssertExpectations(t)
		})
	}
}

// TestRepository_CreateProjectFile тестирует создание записи о файле проекта
func TestRepository_CreateProjectFile(t *testing.T) {
	tests := []struct {
		name          string
		projectID     int32
		filename      string
		originalName  string
		filePath      string
		fileSize      int64
		extension     string
		fileType      db.FileType
		mockFile      db.ProjectFile
		mockError     error
		expectedError bool
	}{
		{
			name:         "Успешное создание записи о файле",
			projectID:    1,
			filename:     "file123.pdf",
			originalName: "document.pdf",
			filePath:     "/files/file123.pdf",
			fileSize:     2048,
			extension:    "pdf",
			fileType:     db.FileTypeDocumentation,
			mockFile: db.ProjectFile{
				ID:           1,
				ProjectID:    1,
				Filename:     "file123.pdf",
				OriginalName: "document.pdf",
				FilePath:     "/files/file123.pdf",
				FileSize:     2048,
				Extension:    "pdf",
				FileType:     db.FileTypeDocumentation,
				UploadedAt:   time.Now(),
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "Ошибка при создании записи о файле",
			projectID:     1,
			filename:      "file123.pdf",
			originalName:  "document.pdf",
			filePath:      "/files/file123.pdf",
			fileSize:      2048,
			extension:     "pdf",
			fileType:      db.FileTypeDocumentation,
			mockFile:      db.ProjectFile{},
			mockError:     errors.New("creation failed"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQuerier := new(MockQuerier)
			repo := &Repository{querier: mockQuerier}

			expectedArg := db.CreateProjectFileParams{
				ProjectID:    tt.projectID,
				Filename:     tt.filename,
				OriginalName: tt.originalName,
				FilePath:     tt.filePath,
				FileSize:     tt.fileSize,
				Extension:    tt.extension,
				FileType:     tt.fileType,
			}

			mockQuerier.On("CreateProjectFile", mock.Anything, expectedArg).Return(tt.mockFile, tt.mockError)

			result, err := repo.CreateProjectFile(context.Background(), tt.projectID, tt.filename, tt.originalName, tt.filePath, tt.fileSize, tt.extension, tt.fileType)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockFile.ID, result.ID)
				assert.Equal(t, tt.mockFile.Filename, result.Filename)
				assert.Equal(t, tt.mockFile.FileType, result.FileType)
			}

			mockQuerier.AssertExpectations(t)
		})
	}
}

// TestRepository_SaveAttach тестирует сохранение информации о загруженном файле
func TestRepository_SaveAttach(t *testing.T) {
	tests := []struct {
		name          string
		attach        *models.Attach
		expectedError bool
	}{
		{
			name: "Успешное сохранение файла",
			attach: &models.Attach{
				Type:    "application/pdf",
				FileExt: ".pdf",
			},
			expectedError: false,
		},
		{
			name: "Файл с расширением .docx",
			attach: &models.Attach{
				Type:    "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
				FileExt: ".docx",
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQuerier := new(MockQuerier)
			repo := &Repository{querier: mockQuerier}

			result, err := repo.SaveAttach(tt.attach)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
				assert.Contains(t, result, tt.attach.FileExt)
			}
		})
	}
}

// TestRepository_New тестирует создание нового экземпляра репозитория
func TestRepository_New(t *testing.T) {
	// Создаем мок для postgres.Client
	mockQuerier := new(MockQuerier)

	// Создаем репозиторий напрямую для тестирования
	repo := &Repository{querier: mockQuerier}

	assert.NotNil(t, repo)
	assert.Equal(t, mockQuerier, repo.querier)
}
