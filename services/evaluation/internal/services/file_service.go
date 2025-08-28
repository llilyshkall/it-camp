package services

import (
	"context"
	"errors"
	"evaluation/internal/models"
	"evaluation/internal/postgres"
	db "evaluation/internal/postgres/sqlc"
	"evaluation/internal/tasks"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// fileService реализация FileService
type fileService struct {
	repo        Repository
	storage     FileStorage
	taskManager tasks.TaskManager
	pgClient    *postgres.Client
}

// NewFileService создает новый экземпляр FileService
func NewFileService(repo Repository, storage FileStorage, taskManager tasks.TaskManager, pgClient *postgres.Client) FileService {
	return &fileService{
		repo:        repo,
		storage:     storage,
		taskManager: taskManager,
		pgClient:    pgClient,
	}
}

// UploadProjectFile загружает файл в проект
func (s *fileService) UploadRemarks(ctx context.Context, projectID int32, file io.Reader, filename, fileType string, fileSize int64) (*db.ProjectFile, error) {
	// Атомарно проверяем статус проекта и изменяем его на "processing_remarks"
	// Если статус не "ready", возвращаем ошибку
	project, err := s.repo.CheckAndUpdateProjectStatus(ctx, projectID, db.ProjectStatusProcessingRemarks)
	if err != nil {
		// Если проект не найден или статус не "ready", возвращаем соответствующую ошибку
		if err.Error() == "no rows in result set" {
			return nil, models.ErrProjectAlreadyProcessing
		}
		return nil, err
	}

	// Логируем успешное изменение статуса проекта
	log.Printf("Project %d status changed to %s for file upload", project.ID, project.Status)

	// Функция для восстановления статуса проекта на 'ready'
	restoreStatus := func() {
		if _, err := s.repo.UpdateProjectStatus(ctx, projectID, db.ProjectStatusReady); err != nil {
			log.Printf("Failed to restore project %d status to ready: %v", projectID, err)
		} else {
			log.Printf("Project %d status restored to ready", projectID)
		}
	}

	// Валидируем тип файла
	var dbFileType db.FileType
	switch fileType {
	case "remarks":
		dbFileType = db.FileTypeRemarks
	default:
		restoreStatus()
		return nil, errors.New("invalid file type")
	}

	// Получаем расширение файла
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		restoreStatus()
		return nil, errors.New("file must have an extension")
	}

	// Определяем MIME тип файла
	contentType := s.getContentType(ext)

	// Генерируем уникальное имя файла
	uniqueFileName := uuid.New().String() + ext

	// Загружаем файл в MinIO
	objectName, err := s.storage.UploadFile(ctx, file, uniqueFileName, contentType)
	if err != nil {
		restoreStatus()
		return nil, err
	}

	// Создаем запись о файле в базе данных
	// Используем objectName как filename, так как это реальное имя файла в MinIO
	projectFile, err := s.repo.CreateProjectFile(ctx, projectID, objectName, filename, objectName, fileSize, ext, dbFileType)
	if err != nil {
		restoreStatus()
		return nil, err
	}

	// Создаем и отправляем задачу ProjectProcessorTask в task manager
	projectTask := tasks.NewProjectProcessorTask(
		projectID,
		1, // Приоритет 1 (высокий)
		s.repo,
		s.storage,
	)

	if err := s.taskManager.SubmitTask(projectTask); err != nil {
		// Логируем ошибку, но не прерываем выполнение
		// TODO: добавить proper logging
		restoreStatus()
		return nil, fmt.Errorf("failed to submit project processing task: %w", err)
	}

	return projectFile, nil
}

// UploadDocumentation загружает файл документации в проект
func (s *fileService) UploadDocumentation(ctx context.Context, projectID int32, file io.Reader, filename string, fileSize int64) (*db.ProjectFile, error) {
	// Проверяем статус проекта - должен быть "ready"
	project, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if project.Status != db.ProjectStatusReady {
		return nil, models.ErrProjectAlreadyProcessing
	}

	// Получаем расширение файла
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return nil, errors.New("file must have an extension")
	}

	// Определяем MIME тип файла
	contentType := s.getContentType(ext)

	// Генерируем уникальное имя файла
	uniqueFileName := uuid.New().String() + ext

	// Загружаем файл в MinIO
	objectName, err := s.storage.UploadFile(ctx, file, uniqueFileName, contentType)
	if err != nil {
		return nil, err
	}

	// Создаем запись о файле в базе данных
	// Используем objectName как filename, так как это реальное имя файла в MinIO
	projectFile, err := s.repo.CreateProjectFile(ctx, projectID, objectName, filename, objectName, fileSize, ext, db.FileTypeDocumentation)
	if err != nil {
		return nil, err
	}

	return projectFile, nil
}

// GenerateChecklist запускает генерацию чеклиста для проекта
func (s *fileService) GenerateChecklist(ctx context.Context, projectID int32) error {
	// Атомарно проверяем статус проекта и изменяем его на "processing_checklist"
	// Если статус не "ready", возвращаем ошибку
	project, err := s.repo.CheckAndUpdateProjectStatus(ctx, projectID, db.ProjectStatusProcessingChecklist)
	if err != nil {
		// Если проект не найден или статус не "ready", возвращаем соответствующую ошибку
		if err.Error() == "no rows in result set" {
			return models.ErrProjectAlreadyProcessing
		}
		return err
	}

	// Логируем успешное изменение статуса проекта
	log.Printf("Project %d status changed to %s for checklist generation", project.ID, project.Status)

	// Создаем и отправляем задачу ProjectProcessorTask в task manager
	projectTask := tasks.NewProjectProcessorTask(
		projectID,
		1, // Приоритет 1 (высокий)
		s.repo,
		s.storage,
	)

	if err := s.taskManager.SubmitTask(projectTask); err != nil {
		// Восстанавливаем статус проекта на 'ready' в случае ошибки
		if _, restoreErr := s.repo.UpdateProjectStatus(ctx, projectID, db.ProjectStatusReady); restoreErr != nil {
			log.Printf("Failed to restore project %d status to ready: %v", projectID, restoreErr)
		} else {
			log.Printf("Project %d status restored to ready after task submission failure", projectID)
		}
		return fmt.Errorf("failed to submit checklist generation task: %w", err)
	}

	return nil
}

// GenerateFinalReport запускает генерацию финального отчета для проекта
func (s *fileService) GenerateFinalReport(ctx context.Context, projectID int32) error {
	// Атомарно проверяем статус проекта и изменяем его на "generating_final_report"
	// Если статус не "ready", возвращаем ошибку
	project, err := s.repo.CheckAndUpdateProjectStatus(ctx, projectID, db.ProjectStatusGeneratingFinalReport)
	if err != nil {
		// Если проект не найден или статус не "ready", возвращаем соответствующую ошибку
		if err.Error() == "no rows in result set" {
			return models.ErrProjectAlreadyProcessing
		}
		return err
	}

	// Логируем успешное изменение статуса проекта
	log.Printf("Project %d status changed to %s for final report generation", project.ID, project.Status)

	// Создаем и отправляем задачу ProjectProcessorTask в task manager
	projectTask := tasks.NewProjectProcessorTask(
		projectID,
		1, // Приоритет 1 (высокий)
		s.repo,
		s.storage,
	)

	if err := s.taskManager.SubmitTask(projectTask); err != nil {
		// Восстанавливаем статус проекта на 'ready' в случае ошибки
		if _, restoreErr := s.repo.UpdateProjectStatus(ctx, projectID, db.ProjectStatusReady); restoreErr != nil {
			log.Printf("Failed to restore project %d status to ready: %v", projectID, restoreErr)
		} else {
			log.Printf("Project %d status restored to ready after task submission failure", projectID)
		}
		return fmt.Errorf("failed to submit final report generation task: %w", err)
	}

	return nil
}

// GetChecklist получает результат проверки чеклиста для проекта
func (s *fileService) GetChecklist(ctx context.Context, projectID int32) (interface{}, error) {
	// Проверяем статус проекта
	project, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Если проект в процессе генерации чеклиста, возвращаем ошибку
	if project.Status == db.ProjectStatusProcessingChecklist {
		return nil, models.ErrChecklistStillGenerating
	}

	// Получаем файлы чеклиста
	files, err := s.repo.GetProjectFilesByType(ctx, projectID, db.FileTypeChecklist)
	if err != nil {
		return nil, err
	}

	// Если файлов нет, возвращаем ошибку
	if len(files) == 0 {
		return nil, models.ErrNotFound404
	}

	// Возвращаем результат
	return map[string]interface{}{
		"project_id": projectID,
		"status":     project.Status,
		"files":      files,
		"message":    "Checklist files found",
	}, nil
}

// GetRemarksClustered получает кластеризированные замечания для проекта
func (s *fileService) GetRemarksClustered(ctx context.Context, projectID int32) (io.ReadCloser, error) {
	// Проверяем статус проекта
	project, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Если проект в процессе обработки замечаний, возвращаем ошибку
	if project.Status == db.ProjectStatusProcessingRemarks {
		return nil, models.ErrRemarksStillProcessing
	}

	// Получаем файлы кластеризированных замечаний
	files, err := s.repo.GetProjectFilesByType(ctx, projectID, db.FileTypeRemarksClustered)
	if err != nil {
		return nil, err
	}

	downloadFile, err := s.storage.DownloadFile(ctx, files[len(files)-1].FilePath)
	if err != nil {
		return nil, err
	}

	return downloadFile, nil
}

// GetFinalReport получает финальный отчет для проекта
func (s *fileService) GetFinalReport(ctx context.Context, projectID int32) (interface{}, error) {
	// Проверяем статус проекта
	project, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Если проект в процессе генерации финального отчета, возвращаем ошибку
	if project.Status == db.ProjectStatusGeneratingFinalReport {
		return nil, models.ErrFinalReportStillGenerating
	}

	// Получаем файлы финального отчета
	files, err := s.repo.GetProjectFilesByType(ctx, projectID, db.FileTypeFinalReport)
	if err != nil {
		return nil, err
	}

	// Если файлов нет, возвращаем ошибку
	if len(files) == 0 {
		return nil, models.ErrNotFound404
	}

	// Возвращаем результат
	return map[string]interface{}{
		"project_id": projectID,
		"status":     project.Status,
		"files":      files,
		"message":    "Final report files found",
	}, nil
}

// getContentType определяет MIME тип файла по расширению
func (s *fileService) getContentType(ext string) string {
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".txt":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}
