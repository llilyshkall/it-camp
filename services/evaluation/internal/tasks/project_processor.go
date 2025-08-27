package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	db "evaluation/internal/postgres/sqlc"
	"evaluation/internal/storage"
	"evaluation/internal/utils"

	"github.com/xuri/excelize/v2"
)

// Repository минимальный интерфейс для репозитория
type Repository interface {
	GetProject(ctx context.Context, id int32) (*db.Project, error)
	GetProjectFilesByType(ctx context.Context, projectID int32, fileType db.FileType) ([]db.ProjectFile, error)
	CreateRemark(ctx context.Context, arg db.CreateRemarkParams) (db.Remark, error)
	CreateProjectFile(ctx context.Context, projectID int32, filename, originalName, filePath string, fileSize int64, extension string, fileType db.FileType) (*db.ProjectFile, error)
}

// RemarkItem структура для элемента замечания из JSON ответа
type RemarkItem struct {
	GroupName          string   `json:"group_name"`
	SynthesizedRemark  string   `json:"synthesized_remark"`
	OriginalDuplicates []string `json:"original_duplicates"`
}

// RemarksResponse структура для JSON ответа от внешнего сервиса
type RemarksResponse map[string][]RemarkItem

// ProjectProcessorTask задача для обработки проекта
type ProjectProcessorTask struct {
	projectID int32
	priority  int
	repo      Repository
	storage   storage.FileStorage
}

// NewProjectProcessorTask создает новую задачу обработки проекта
func NewProjectProcessorTask(
	projectID int32,
	priority int,
	repo Repository,
	storage storage.FileStorage,
) *ProjectProcessorTask {
	return &ProjectProcessorTask{
		projectID: projectID,
		priority:  priority,
		repo:      repo,
		storage:   storage,
	}
}

// Execute выполняет задачу обработки проекта
func (pt *ProjectProcessorTask) Execute(ctx context.Context) error {
	log.Printf("Starting project processing task for project %d", pt.projectID)

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
	project, err := pt.repo.GetProject(ctx, pt.projectID)
	if err != nil {
		return nil, err
	}

	return project, nil
}

// processRemarks обрабатывает замечания проекта
func (pt *ProjectProcessorTask) processRemarks(ctx context.Context, project *db.Project) error {
	log.Printf("Processing remarks for project %d", pt.projectID)

	// Имитируем обработку замечаний
	time.Sleep(2 * time.Second)

	// TODO: process remarks
	//
	files, err := pt.repo.GetProjectFilesByType(ctx, project.ID, db.FileTypeRemarks)
	if err != nil || len(files) == 0 {
		return err
	}

	fileRemarks := files[0]

	// Получаем файл из S3
	fileReader, err := pt.storage.DownloadFile(ctx, fileRemarks.Filename)
	if err != nil {
		return fmt.Errorf("failed to download file %s from S3: %w", fileRemarks.Filename, err)
	}
	defer fileReader.Close()

	// Читаем содержимое файла
	fileContent, err := io.ReadAll(fileReader)
	if err != nil {
		return fmt.Errorf("failed to read file content: %w", err)
	}

	log.Printf("Successfully downloaded file %s from S3, size: %d bytes", fileRemarks.Filename, len(fileContent))

	// Парсим Excel файл и преобразуем в JSON
	jsonData, err := utils.ParseExcelFromBytes(fileContent)
	if err != nil {
		return fmt.Errorf("failed to parse Excel file: %w", err)
	}

	externalURL := "http://127.0.0.1:8083/remarks"

	// Send request to external service
	resp, err := http.Post(externalURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send remarks to external service: %v", err)
	}
	defer resp.Body.Close()

	// Check external service response
	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("external service returned status %d: %s", resp.StatusCode, string(body))
	}

	// Читаем ответ от внешнего сервиса
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Парсим JSON ответ
	var remarksResponse RemarksResponse
	if err := json.Unmarshal(respBody, &remarksResponse); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Сохраняем замечания в БД
	if err := pt.saveRemarksToDB(ctx, project.ID, remarksResponse); err != nil {
		return fmt.Errorf("failed to save remarks to DB: %w", err)
	}

	log.Printf("Successfully saved %d remark categories to DB", len(remarksResponse))

	// Генерируем Excel файл из JSON ответа
	excelBuffer, err := pt.generateExcelFromRemarks(remarksResponse)
	if err != nil {
		return fmt.Errorf("failed to generate Excel file: %w", err)
	}

	// Сохраняем Excel файл в S3
	objectName, err := pt.storage.UploadFile(ctx, excelBuffer, "remarks_clustered.xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	if err != nil {
		return fmt.Errorf("failed to upload Excel file to S3: %w", err)
	}

	// Сохраняем запись о файле в БД
	_, err = pt.repo.CreateProjectFile(ctx, project.ID, "remarks_clustered.xlsx", "Замечания по тематикам.xlsx", objectName, int64(excelBuffer.Len()), ".xlsx", db.FileTypeRemarksClustered)
	if err != nil {
		return fmt.Errorf("failed to create project file record: %w", err)
	}

	log.Printf("Successfully generated and uploaded Excel file %s to S3", objectName)

	log.Printf("Successfully processed remarks for project %d", pt.projectID)
	return nil
}

// generateChecklist генерирует чек-лист для проекта
func (pt *ProjectProcessorTask) generateChecklist(ctx context.Context, project *db.Project) error {
	log.Printf("Generating checklist for project %d", pt.projectID)

	// Имитируем генерацию чек-листа
	time.Sleep(3 * time.Second)

	// TODO: generate checklist

	log.Printf("Successfully generated checklist for project %d", pt.projectID)
	return nil
}

// generateFinalReport генерирует итоговый отчет
func (pt *ProjectProcessorTask) generateFinalReport(ctx context.Context, project *db.Project) error {
	log.Printf("Generating final report for project %d", pt.projectID)

	// Имитируем генерацию отчета
	time.Sleep(4 * time.Second)

	// TODO: generate final report

	log.Printf("Successfully generated final report for project %d", pt.projectID)
	return nil
}

// saveRemarksToDB сохраняет замечания в базу данных
func (pt *ProjectProcessorTask) saveRemarksToDB(ctx context.Context, projectID int32, remarksResponse RemarksResponse) error {
	for section, items := range remarksResponse {
		for _, item := range items {
			// Создаем замечание в БД
			_, err := pt.repo.CreateRemark(ctx, db.CreateRemarkParams{
				ProjectID:  projectID,
				Direction:  "",
				Section:    section,
				Subsection: item.GroupName, // Пока оставляем пустым, можно добавить логику для подразделов
				Content:    item.SynthesizedRemark,
			})
			if err != nil {
				return fmt.Errorf("failed to create remark for direction %s, section %s: %w", section, item.GroupName, err)
			}
		}
	}
	return nil
}

// generateExcelFromRemarks генерирует Excel файл из замечаний
func (pt *ProjectProcessorTask) generateExcelFromRemarks(remarksResponse RemarksResponse) (*bytes.Buffer, error) {
	// Создаем новый Excel файл
	f := excelize.NewFile()
	defer f.Close()

	// Создаем лист для замечаний
	sheetName := "Замечания"
	f.NewSheet(sheetName)

	// Устанавливаем заголовки
	headers := []string{"Раздел", "Подраздел", "Синтезированное замечание", "Оригинальные замечания"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	// Заполняем данными
	row := 2
	for section, items := range remarksResponse {
		for _, item := range items {
			// Раздел
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), section)
			// Подраздел
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), item.GroupName)
			// Синтезированное замечание
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), item.SynthesizedRemark)
			// Оригинальные замечания (объединяем в одну строку)
			originalRemarks := ""
			for i, remark := range item.OriginalDuplicates {
				if i > 0 {
					originalRemarks += "; "
				}
				originalRemarks += remark
			}
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), originalRemarks)
			row++
		}
	}

	// Автоматически подгоняем ширину столбцов
	for i := 0; i < len(headers); i++ {
		col := string(rune('A' + i))
		f.SetColWidth(sheetName, col, col, 30)
	}

	// Сохраняем в буфер
	buffer := new(bytes.Buffer)
	if err := f.Write(buffer); err != nil {
		return nil, fmt.Errorf("failed to write Excel file to buffer: %w", err)
	}

	return buffer, nil
}
