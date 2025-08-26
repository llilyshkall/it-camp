package tasks

import (
	"context"
	"fmt"
	"log"
	"time"

	"evaluation/internal/postgres"
	db "evaluation/internal/postgres/sqlc"
	"evaluation/internal/storage"
)

// ProjectProcessorTask задача для обработки проекта
type ProjectProcessorTask struct {
	projectID int32
	priority  int
	pgClient  *postgres.Client
	storage   storage.FileStorage
}

// NewProjectProcessorTask создает новую задачу обработки проекта
func NewProjectProcessorTask(
	projectID int32,
	priority int,
	pgClient *postgres.Client,
	storage storage.FileStorage,
) *ProjectProcessorTask {
	return &ProjectProcessorTask{
		projectID: projectID,
		priority:  priority,
		pgClient:  pgClient,
		storage:   storage,
	}
}

// Execute выполняет задачу обработки проекта
func (pt *ProjectProcessorTask) Execute(ctx context.Context) error {
	log.Printf("Starting project processing task for project %d, type: %s", pt.projectID)

	// Получаем информацию о проекте
	project, err := pt.getProject(ctx)
	if err != nil {
		return fmt.Errorf("failed to get project %d: %w", pt.projectID, err)
	}

	// Выполняем обработку в зависимости от типа задачи
	switch project.Status {
	case db.ProjectStatusProcessingRemarks:
		return pt.processRemarks(ctx, project)
	case db.ProjectStatusProcessingChecklist:
		return pt.generateChecklist(ctx, project)
	case db.ProjectStatusGeneratingFinalReport:
		return pt.generateFinalReport(ctx, project)
	default:
		return fmt.Errorf("unknown task type: %s", project.Status)
	}
}

// GetProjectID возвращает ID проекта
func (pt *ProjectProcessorTask) GetProjectID() int32 {
	return pt.projectID
}

// GetPriority возвращает приоритет задачи
func (pt *ProjectProcessorTask) GetPriority() int {
	return pt.priority
}

// getProject получает информацию о проекте из БД
func (pt *ProjectProcessorTask) getProject(ctx context.Context) (*db.Project, error) {
	querier := db.New(pt.pgClient.DB)

	project, err := querier.GetProject(ctx, pt.projectID)
	if err != nil {
		return nil, err
	}

	return &project, nil
}

// getProjectFiles получает файлы проекта из БД
func (pt *ProjectProcessorTask) getProjectFiles(ctx context.Context) ([]db.ProjectFile, error) {
	querier := db.New(pt.pgClient.DB)

	files, err := querier.GetProjectFiles(ctx, pt.projectID)
	if err != nil {
		return nil, err
	}

	return files, nil
}

// processRemarks обрабатывает замечания проекта
func (pt *ProjectProcessorTask) processRemarks(ctx context.Context, project *db.Project) error {
	log.Printf("Processing remarks for project %d", pt.projectID)

	// Имитируем обработку замечаний
	time.Sleep(2 * time.Second)

	// TODO: process remarks

	// Обновляем статус проекта
	querier := db.New(pt.pgClient.DB)
	_, err := querier.UpdateProjectStatus(ctx, db.UpdateProjectStatusParams{
		ID:     pt.projectID,
		Status: db.ProjectStatusProcessingChecklist,
	})
	if err != nil {
		return fmt.Errorf("failed to update project status: %w", err)
	}

	log.Printf("Successfully processed remarks for project %d", pt.projectID)
	return nil
}

// generateChecklist генерирует чек-лист для проекта
func (pt *ProjectProcessorTask) generateChecklist(ctx context.Context, project *db.Project) error {
	log.Printf("Generating checklist for project %d", pt.projectID)

	// Имитируем генерацию чек-листа
	time.Sleep(3 * time.Second)

	// TODO: generate checklist

	// Обновляем статус проекта
	querier := db.New(pt.pgClient.DB)
	_, err := querier.UpdateProjectStatus(ctx, db.UpdateProjectStatusParams{
		ID:     pt.projectID,
		Status: db.ProjectStatusGeneratingFinalReport,
	})
	if err != nil {
		return fmt.Errorf("failed to update project status: %w", err)
	}

	log.Printf("Successfully generated checklist for project %d", pt.projectID)
	return nil
}

// generateFinalReport генерирует итоговый отчет
func (pt *ProjectProcessorTask) generateFinalReport(ctx context.Context, project *db.Project) error {
	log.Printf("Generating final report for project %d", pt.projectID)

	// Имитируем генерацию отчета
	time.Sleep(4 * time.Second)

	// TODO: generate final report

	// Обновляем статус проекта на завершенный
	querier := db.New(pt.pgClient.DB)
	_, err := querier.UpdateProjectStatus(ctx, db.UpdateProjectStatusParams{
		ID:     pt.projectID,
		Status: db.ProjectStatusReady,
	})
	if err != nil {
		return fmt.Errorf("failed to update project status: %w", err)
	}

	log.Printf("Successfully generated final report for project %d", pt.projectID)
	return nil
}
