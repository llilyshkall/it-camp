package services

import (
	"context"
	"evaluation/internal/models"
	db "evaluation/internal/postgres/sqlc"
	"io"
)

// Repository интерфейс для репозитория
type Repository interface {
	CreateProject(ctx context.Context, name string) (*db.Project, error)
	GetProject(ctx context.Context, id int32) (*db.Project, error)
	ListProjects(ctx context.Context) ([]db.Project, error)
	CreateProjectFile(ctx context.Context, projectID int32, filename, originalName, filePath string, fileSize int64, extension string, fileType db.FileType) (*db.ProjectFile, error)
	CheckAndUpdateProjectStatus(ctx context.Context, projectID int32, newStatus db.ProjectStatus) (*db.Project, error)
	UpdateProjectStatus(ctx context.Context, projectID int32, newStatus db.ProjectStatus) (*db.Project, error)
	GetProjectFilesByType(ctx context.Context, projectID int32, fileType db.FileType) ([]db.ProjectFile, error)
	CreateRemark(ctx context.Context, arg db.CreateRemarkParams) (db.Remark, error)
	SaveAttach(file *models.Attach) (string, error)
}

// FileStorage интерфейс для файлового хранилища
type FileStorage interface {
	UploadFile(ctx context.Context, file io.Reader, filename, contentType string) (string, error)
	DownloadFile(ctx context.Context, objectName string) (io.ReadCloser, error)
	DeleteFile(ctx context.Context, objectName string) error
	GetFileURL(objectName string) string
}

// ProjectService интерфейс для бизнес-логики проектов
type ProjectService interface {
	CreateProject(ctx context.Context, name string) (*db.Project, error)
	GetProject(ctx context.Context, id int32) (*db.Project, error)
	ListProjects(ctx context.Context) ([]db.Project, error)
}

// FileService интерфейс для бизнес-логики файлов
type FileService interface {
	UploadRemarks(ctx context.Context, projectID int32, file io.Reader, filename, fileType string, fileSize int64) (*db.ProjectFile, error)
	UploadDocumentation(ctx context.Context, projectID int32, file io.Reader, filename string, fileSize int64) (*db.ProjectFile, error)
	GenerateChecklist(ctx context.Context, projectID int32) error
	GenerateFinalReport(ctx context.Context, projectID int32) error
	GetChecklist(ctx context.Context, projectID int32) (interface{}, error)
	GetRemarksClustered(ctx context.Context, projectID int32) (interface{}, error)
	GetFinalReport(ctx context.Context, projectID int32) (interface{}, error)
}

// HealthService интерфейс для проверки состояния сервиса
type HealthService interface {
	CheckHealth(ctx context.Context) (*HealthResponse, error)
}

// HealthResponse структура ответа для проверки состояния
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Database  string `json:"database"`
}
