package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	db "evaluation/internal/postgres/sqlc"
	"evaluation/internal/storage"
	"evaluation/internal/utils"

	"github.com/go-resty/resty/v2"
	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

// RAGConfig конфигурация для RAG-системы
type RAGConfig struct {
	LLMAPIURL    string
	LLMModelName string
	MaxChunkSize int
	ChunkOverlap int
	TopK         int
	RequestDelay time.Duration
}

// DocumentChunk чанк документа для индексации
type DocumentChunk struct {
	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata"`
}

// ChecklistItem элемент чек-листа
type ChecklistItem struct {
	Criterion string `json:"criterion"`
	Status    string `json:"status"`
	Answer    string `json:"answer"`
	Sources   []struct {
		Filename string `json:"filename"`
		Page     string `json:"page"`
		Snippet  string `json:"snippet"`
	} `json:"sources"`
}

// LLMResponse ответ от LLM API
type LLMResponse struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

// ExternalServiceResponse ответ от внешнего сервиса замечаний
type ExternalServiceResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Development            []RemarkItem `json:"development"`
		Geological            []RemarkItem `json:"geological"`
		HydrodynamicIntegrated []RemarkItem `json:"hydrodynamic_integrated"`
		Petrophysical         []RemarkItem `json:"petrophysical"`
		Reassessment          []RemarkItem `json:"reassessment"`
		Seismogeological      []RemarkItem `json:"seismogeological"`
	} `json:"data"`
}

// RAGSystem система для RAG-операций
type RAGSystem struct {
	config    RAGConfig
	client    *resty.Client
	documents []DocumentChunk
}

// NewRAGSystem создает новую RAG-систему
func NewRAGSystem(config RAGConfig) *RAGSystem {
	client := resty.New().
		SetTimeout(300 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second)

	return &RAGSystem{
		config:    config,
		client:    client,
		documents: []DocumentChunk{},
	}
}

// extractTextFromFile извлекает текст из файла по его расширению
func (rag *RAGSystem) extractTextFromFile(file io.Reader, filename string) (string, error) {
	// Простая реализация для текстовых файлов
	// В реальном проекте здесь можно добавить парсинг PDF, DOCX и других форматов
	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Убираем HTML теги если есть
	contentStr := string(content)
	re := regexp.MustCompile(`<[^>]*>`)
	contentStr = re.ReplaceAllString(contentStr, "")

	// Убираем лишние пробелы
	contentStr = strings.TrimSpace(contentStr)

	return contentStr, nil
}

// splitTextIntoChunks разбивает текст на чанки
func (rag *RAGSystem) splitTextIntoChunks(text string, filename string) []DocumentChunk {
	var chunks []DocumentChunk

	// Простое разбиение по предложениям
	sentences := regexp.MustCompile(`[.!?]+`).Split(text, -1)

	for i, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) > 10 { // Минимальная длина предложения
			chunk := DocumentChunk{
				Content: sentence,
				Metadata: map[string]string{
					"filename": filename,
					"chunk_id": fmt.Sprintf("%d", i),
				},
			}
			chunks = append(chunks, chunk)
		}
	}

	return chunks
}

// searchRelevantChunks ищет релевантные чанки по запросу
func (rag *RAGSystem) searchRelevantChunks(query string) []DocumentChunk {
	var relevantChunks []DocumentChunk

	// Простой поиск по ключевым словам
	queryLower := strings.ToLower(query)
	queryWords := strings.Fields(queryLower)

	for _, chunk := range rag.documents {
		chunkLower := strings.ToLower(chunk.Content)
		score := 0

		for _, word := range queryWords {
			if strings.Contains(chunkLower, word) {
				score++
			}
		}

		if score > 0 {
			relevantChunks = append(relevantChunks, chunk)
		}
	}

	// Ограничиваем количество результатов
	if len(relevantChunks) > rag.config.TopK {
		relevantChunks = relevantChunks[:rag.config.TopK]
	}

	return relevantChunks
}

// callLLM отправляет запрос к LLM API
func (rag *RAGSystem) callLLM(messages []map[string]string) (string, error) {
	payload := map[string]interface{}{
		"model":    rag.config.LLMModelName,
		"messages": messages,
		"stream":   false,
		"format":   "json",
		"options": map[string]interface{}{
			"temperature":        0.2,
			"top_p":              0.9,
			"repetition_penalty": 1.05,
		},
	}

	resp, err := rag.client.R().
		SetBody(payload).
		SetResult(&LLMResponse{}).
		Post(rag.config.LLMAPIURL)

	if err != nil {
		return "", fmt.Errorf("failed to call LLM API: %w", err)
	}

	if resp.StatusCode() != 200 {
		return "", fmt.Errorf("LLM API returned status %d: %s", resp.StatusCode(), resp.String())
	}

	llmResp := resp.Result().(*LLMResponse)
	return llmResp.Message.Content, nil
}

// processCriterion обрабатывает один критерий чек-листа
func (rag *RAGSystem) processCriterion(criterion string) (*ChecklistItem, error) {
	// Ищем релевантные документы
	relevantChunks := rag.searchRelevantChunks(criterion)

	if len(relevantChunks) == 0 {
		return &ChecklistItem{
			Criterion: criterion,
			Status:    "not_found",
			Answer:    "Не найдено релевантных документов.",
			Sources: []struct {
				Filename string `json:"filename"`
				Page     string `json:"page"`
				Snippet  string `json:"snippet"`
			}{},
		}, nil
	}

	// Формируем контекст для LLM
	var contextBuilder strings.Builder
	for i, chunk := range relevantChunks {
		contextBuilder.WriteString(fmt.Sprintf("[ИСТОЧНИК %d: %s]\n", i+1, chunk.Metadata["filename"]))
		contextBuilder.WriteString(chunk.Content)
		contextBuilder.WriteString("\n\n")
	}

	// Формируем промпт для LLM
	prompt := fmt.Sprintf(`Ты — ассистент-аналитик, который возвращает ответы строго в формате JSON. Проанализируй предоставленный КОНТЕКСТ и ответь на ВОПРОС НА РУССКОМ.

КОНТЕКСТ:
---
%s
---

ВОПРОС: "%s"

Твой ответ должен быть ТОЛЬКО JSON объектом со следующей структурой:
{
  "status": "ОДИН ИЗ СТАТУСОВ: confirmed, not_found, partial, indirect, requires_confirmation",
  "answer": "Твой развернутый ответ на основе контекста, со ссылками на источники в формате [ИСТОЧНИК N] НА РУССКОМ"
}`, contextBuilder.String(), criterion)

	// Отправляем запрос к LLM
	response, err := rag.callLLM([]map[string]string{
		{"role": "user", "content": prompt},
	})

	if err != nil {
		return &ChecklistItem{
			Criterion: criterion,
			Status:    "requires_confirmation",
			Answer:    fmt.Sprintf("Ошибка при обращении к LLM: %v", err),
			Sources: []struct {
				Filename string `json:"filename"`
				Page     string `json:"page"`
				Snippet  string `json:"snippet"`
			}{},
		}, nil
	}

	// Парсим ответ LLM
	var llmResult struct {
		Status string `json:"status"`
		Answer string `json:"answer"`
	}

	if err := json.Unmarshal([]byte(response), &llmResult); err != nil {
		llmResult.Status = "requires_confirmation"
		llmResult.Answer = fmt.Sprintf("Ошибка парсинга ответа LLM: %v. Ответ: %s", err, response)
	}

	// Формируем источники
	var sources []struct {
		Filename string `json:"filename"`
		Page     string `json:"page"`
		Snippet  string `json:"snippet"`
	}

	for _, chunk := range relevantChunks {
		sources = append(sources, struct {
			Filename string `json:"filename"`
			Page     string `json:"page"`
			Snippet  string `json:"snippet"`
		}{
			Filename: chunk.Metadata["filename"],
			Page:     chunk.Metadata["chunk_id"],
			Snippet:  chunk.Content,
		})
	}

	return &ChecklistItem{
		Criterion: criterion,
		Status:    llmResult.Status,
		Answer:    llmResult.Answer,
		Sources:   sources,
	}, nil
}

// Repository минимальный интерфейс для репозитория
type Repository interface {
	GetProject(ctx context.Context, id int32) (*db.Project, error)
	GetProjectFilesByType(ctx context.Context, projectID int32, fileType db.FileType) ([]db.ProjectFile, error)
	CreateRemark(ctx context.Context, arg db.CreateRemarkParams) (db.Remark, error)
	CreateProjectFile(ctx context.Context, projectID int32, filename, originalName, filePath string, fileSize int64, extension string, fileType db.FileType) (*db.ProjectFile, error)
	UpdateProjectStatus(ctx context.Context, projectID int32, newStatus db.ProjectStatus) (*db.Project, error)
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
	//querier := db.New(pt.pgClient.DB)
	// Имитируем обработку замечаний
	//time.Sleep(2 * time.Second)

	// TODO: process remarks
	//
	files, err := pt.repo.GetProjectFilesByType(ctx, project.ID, db.FileTypeRemarks)
	if err != nil || len(files) == 0 {
		// Устанавливаем статус ready при ошибке
		if updateErr := pt.setProjectStatusReady(ctx, project.ID); updateErr != nil {
			log.Printf("Failed to set project status to ready after error: %v", updateErr)
		}
		return err
	}

	fileRemarks := files[0]

	// Получаем файл из S3
	fileReader, err := pt.storage.DownloadFile(ctx, fileRemarks.Filename)
	if err != nil {
		// Устанавливаем статус ready при ошибке
		if updateErr := pt.setProjectStatusReady(ctx, project.ID); updateErr != nil {
			log.Printf("Failed to set project status to ready after error: %v", updateErr)
		}
		return fmt.Errorf("failed to download file %s from S3: %w", fileRemarks.Filename, err)
	}
	defer fileReader.Close()

	// Читаем содержимое файла
	fileContent, err := io.ReadAll(fileReader)
	if err != nil {
		// Устанавливаем статус ready при ошибке
		if updateErr := pt.setProjectStatusReady(ctx, project.ID); updateErr != nil {
			log.Printf("Failed to set project status to ready after error: %v", updateErr)
		}
		return fmt.Errorf("failed to read file content: %w", err)
	}

	log.Printf("Successfully downloaded file %s from S3, size: %d bytes", fileRemarks.Filename, len(fileContent))

	// Парсим Excel файл и преобразуем в JSON
	jsonData, err := utils.ParseExcelFromBytes(fileContent)
	if err != nil {
		// Устанавливаем статус ready при ошибке
		if updateErr := pt.setProjectStatusReady(ctx, project.ID); updateErr != nil {
			log.Printf("Failed to set project status to ready after error: %v", updateErr)
		}
		return fmt.Errorf("failed to parse Excel file: %w", err)
	}

	externalURL := "http://127.0.0.1:8083/remarks"

	// Send request to external service
	resp, err := http.Post(externalURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		// Устанавливаем статус ready при ошибке
		if updateErr := pt.setProjectStatusReady(ctx, project.ID); updateErr != nil {
			log.Printf("Failed to set project status to ready after error: %v", updateErr)
		}
		return fmt.Errorf("failed to send remarks to external service: %v", err)
	}
	defer resp.Body.Close()

	// Check external service response
	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		// Устанавливаем статус ready при ошибке
		if updateErr := pt.setProjectStatusReady(ctx, project.ID); updateErr != nil {
			log.Printf("Failed to set project status to ready after error: %v", updateErr)
		}
		return fmt.Errorf("external service returned status %d: %s", resp.StatusCode, string(body))
	}

	// Читаем ответ от внешнего сервиса
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		// Устанавливаем статус ready при ошибке
		if updateErr := pt.setProjectStatusReady(ctx, project.ID); updateErr != nil {
			log.Printf("Failed to set project status to ready after error: %v", updateErr)
		}
		return fmt.Errorf("failed to read response body: %w", err)
	}
	log.Println(string(respBody))

	// Парсим JSON ответ от внешнего сервиса
	var externalResponse ExternalServiceResponse
	if err := json.Unmarshal(respBody, &externalResponse); err != nil {
		// Устанавливаем статус ready при ошибке
		if updateErr := pt.setProjectStatusReady(ctx, project.ID); updateErr != nil {
			log.Printf("Failed to set project status to ready after error: %v", updateErr)
		}
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Проверяем успешность ответа
	if !externalResponse.Success {
		// Устанавливаем статус ready при ошибке
		if updateErr := pt.setProjectStatusReady(ctx, project.ID); updateErr != nil {
			log.Printf("Failed to set project status to ready after error: %v", updateErr)
		}
		return fmt.Errorf("external service returned unsuccessful response")
	}

	// Преобразуем ответ в формат RemarksResponse для совместимости
	remarksResponse := make(RemarksResponse)
	remarksResponse["development"] = externalResponse.Data.Development
	remarksResponse["geological"] = externalResponse.Data.Geological
	remarksResponse["hydrodynamic_integrated"] = externalResponse.Data.HydrodynamicIntegrated
	remarksResponse["petrophysical"] = externalResponse.Data.Petrophysical
	remarksResponse["reassessment"] = externalResponse.Data.Reassessment
	remarksResponse["seismogeological"] = externalResponse.Data.Seismogeological

	// Сохраняем замечания в БД
	if err := pt.saveRemarksToDB(ctx, project.ID, remarksResponse); err != nil {
		// Устанавливаем статус ready при ошибке
		if updateErr := pt.setProjectStatusReady(ctx, project.ID); updateErr != nil {
			log.Printf("Failed to set project status to ready after error: %v", updateErr)
		}
		return fmt.Errorf("failed to save remarks to DB: %w", err)
	}

	log.Printf("Successfully saved %d remark categories to DB", len(remarksResponse))

	// Генерируем PDF отчет из JSON ответа
	pdfBuffer, err := pt.generatePDFFromRemarks(remarksResponse)
	if err != nil {
		// Устанавливаем статус ready при ошибке
		if updateErr := pt.setProjectStatusReady(ctx, project.ID); updateErr != nil {
			log.Printf("Failed to set project status to ready after error: %v", updateErr)
		}
		return fmt.Errorf("failed to generate PDF report: %w", err)
	}

	// Сохраняем PDF файл в S3
	objectName, err := pt.storage.UploadFile(ctx, pdfBuffer, "remarks_report.pdf", "application/pdf")
	if err != nil {
		// Устанавливаем статус ready при ошибке
		if updateErr := pt.setProjectStatusReady(ctx, project.ID); updateErr != nil {
			log.Printf("Failed to set project status to ready after error: %v", updateErr)
		}
		return fmt.Errorf("failed to upload PDF file to S3: %w", err)
	}

	// Сохраняем запись о файле в БД
	_, err = pt.repo.CreateProjectFile(ctx, project.ID, "remarks_report.pdf", "Отчет по замечаниям.pdf", objectName, int64(pdfBuffer.Len()), ".pdf", db.FileTypeRemarksClustered)
	if err != nil {
		// Устанавливаем статус ready при ошибке
		if updateErr := pt.setProjectStatusReady(ctx, project.ID); updateErr != nil {
			log.Printf("Failed to set project status to ready after error: %v", updateErr)
		}
		return fmt.Errorf("failed to create project file record: %w", err)
	}

	log.Printf("Successfully generated and uploaded PDF report %s to S3", objectName)

	log.Printf("Successfully processed remarks for project %d", pt.projectID)

	// Устанавливаем статус ready после успешной обработки
	if err := pt.setProjectStatusReady(ctx, project.ID); err != nil {
		log.Printf("Failed to set project status to ready: %v", err)
		return fmt.Errorf("failed to set project status to ready: %w", err)
	}

	return nil
}

// setProjectStatusReady устанавливает статус проекта на ready
func (pt *ProjectProcessorTask) setProjectStatusReady(ctx context.Context, projectID int32) error {
	_, err := pt.repo.UpdateProjectStatus(ctx, projectID, db.ProjectStatusReady)
	if err != nil {
		return fmt.Errorf("failed to update project status to ready: %w", err)
	}
	log.Printf("Project %d status set to ready", projectID)
	return nil
}

// generateChecklist генерирует чек-лист для проекта
func (pt *ProjectProcessorTask) generateChecklist(ctx context.Context, project *db.Project) error {
	log.Printf("Generating checklist for project %d", pt.projectID)

	// Получаем файлы документации для проекта
	docFiles, err := pt.repo.GetProjectFilesByType(ctx, pt.projectID, db.FileTypeDocumentation)
	if err != nil {
		return fmt.Errorf("failed to get documentation files: %w", err)
	}

	if len(docFiles) == 0 {
		log.Printf("No documentation files found for project %d", pt.projectID)
		// Устанавливаем статус ready если нет файлов для обработки
		return pt.setProjectStatusReady(ctx, project.ID)
	}

	// Создаем RAG-систему
	ragConfig := RAGConfig{
		LLMAPIURL:    "http://89.108.116.240:11434/api/chat",
		LLMModelName: "qwen3-8b:latest",
		MaxChunkSize: 700,
		ChunkOverlap: 150,
		TopK:         5,
		RequestDelay: 500 * time.Millisecond,
	}

	rag := NewRAGSystem(ragConfig)

	// Обрабатываем каждый файл документации
	for _, docFile := range docFiles {
		log.Printf("Processing documentation file: %s", docFile.Filename)

		// Скачиваем файл из S3
		fileReader, err := pt.storage.DownloadFile(ctx, docFile.FilePath)
		if err != nil {
			log.Printf("Failed to download file %s: %v", docFile.Filename, err)
			continue
		}
		defer fileReader.Close()

		// Извлекаем текст из файла
		text, err := rag.extractTextFromFile(fileReader, docFile.OriginalName)
		if err != nil {
			log.Printf("Failed to extract text from file %s: %v", docFile.Filename, err)
			continue
		}

		// Разбиваем на чанки и добавляем в RAG-систему
		chunks := rag.splitTextIntoChunks(text, docFile.OriginalName)
		rag.documents = append(rag.documents, chunks...)

		log.Printf("Added %d chunks from file %s", len(chunks), docFile.Filename)
	}

	// Получаем чек-лист из CSV файла (если есть)
	checklistFile, err := pt.repo.GetProjectFilesByType(ctx, pt.projectID, db.FileTypeChecklist)
	if err != nil {
		log.Printf("Failed to get checklist file: %v", err)
		// Создаем базовый чек-лист
		return pt.createBasicChecklist(ctx, project, rag)
	}

	if len(checklistFile) == 0 {
		log.Printf("No checklist file found, creating basic checklist")
		return pt.createBasicChecklist(ctx, project, rag)
	}

	// Обрабатываем чек-лист
	return pt.processChecklistFile(ctx, project, checklistFile[0], rag)
}

// createBasicChecklist создает базовый чек-лист для проекта
func (pt *ProjectProcessorTask) createBasicChecklist(ctx context.Context, project *db.Project, rag *RAGSystem) error {
	// Базовые критерии для проверки
	basicCriteria := []string{
		"Наличие технического задания",
		"Наличие проектной документации",
		"Наличие исполнительной документации",
		"Соответствие требованиям безопасности",
		"Соответствие нормативным требованиям",
	}

	var checklistResults []ChecklistItem

	// Обрабатываем каждый критерий
	for _, criterion := range basicCriteria {
		log.Printf("Processing criterion: %s", criterion)

		result, err := rag.processCriterion(criterion)
		if err != nil {
			log.Printf("Failed to process criterion '%s': %v", criterion, err)
			result = &ChecklistItem{
				Criterion: criterion,
				Status:    "requires_confirmation",
				Answer:    fmt.Sprintf("Ошибка обработки: %v", err),
				Sources: []struct {
					Filename string `json:"filename"`
					Page     string `json:"page"`
					Snippet  string `json:"snippet"`
				}{},
			}
		}

		checklistResults = append(checklistResults, *result)

		// Небольшая задержка между запросами
		time.Sleep(rag.config.RequestDelay)
	}

	// Сохраняем результаты в JSON файл
	return pt.saveChecklistResults(ctx, project, checklistResults, "basic_checklist")
}

// processChecklistFile обрабатывает файл чек-листа
func (pt *ProjectProcessorTask) processChecklistFile(ctx context.Context, project *db.Project, checklistFile db.ProjectFile, rag *RAGSystem) error {
	// Скачиваем файл чек-листа
	fileReader, err := pt.storage.DownloadFile(ctx, checklistFile.FilePath)
	if err != nil {
		return fmt.Errorf("failed to download checklist file: %w", err)
	}
	defer fileReader.Close()

	// Читаем содержимое файла
	content, err := io.ReadAll(fileReader)
	if err != nil {
		return fmt.Errorf("failed to read checklist file: %w", err)
	}

	// Парсим CSV (простая реализация)
	lines := strings.Split(string(content), "\n")
	var criteria []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "criterion") {
			// Простое извлечение критерия (в реальном проекте лучше использовать CSV парсер)
			if strings.Contains(line, ",") {
				parts := strings.Split(line, ",")
				if len(parts) > 0 {
					criteria = append(criteria, strings.TrimSpace(parts[0]))
				}
			} else {
				criteria = append(criteria, line)
			}
		}
	}

	if len(criteria) == 0 {
		log.Printf("No criteria found in checklist file")
		return pt.createBasicChecklist(ctx, project, rag)
	}

	var checklistResults []ChecklistItem

	// Обрабатываем каждый критерий
	for _, criterion := range criteria {
		log.Printf("Processing criterion: %s", criterion)

		result, err := rag.processCriterion(criterion)
		if err != nil {
			log.Printf("Failed to process criterion '%s': %v", criterion, err)
			result = &ChecklistItem{
				Criterion: criterion,
				Status:    "requires_confirmation",
				Answer:    fmt.Sprintf("Ошибка обработки: %v", err),
				Sources: []struct {
					Filename string `json:"filename"`
					Page     string `json:"page"`
					Snippet  string `json:"snippet"`
				}{},
			}
		}

		checklistResults = append(checklistResults, *result)

		// Небольшая задержка между запросами
		time.Sleep(rag.config.RequestDelay)
	}

	// Сохраняем результаты
	return pt.saveChecklistResults(ctx, project, checklistResults, "checklist_verification")
}

// saveChecklistResults сохраняет результаты проверки чек-листа
func (pt *ProjectProcessorTask) saveChecklistResults(ctx context.Context, project *db.Project, results []ChecklistItem, reportType string) error {
	// Создаем JSON отчет
	reportData := map[string]interface{}{
		"project_id":   project.ID,
		"project_name": project.Name,
		"report_type":  reportType,
		"generated_at": time.Now().Format(time.RFC3339),
		"results":      results,
	}

	reportJSON, err := json.MarshalIndent(reportData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	// Сохраняем JSON отчет в S3
	reportBuffer := bytes.NewBuffer(reportJSON)
	objectName, err := pt.storage.UploadFile(ctx, reportBuffer, fmt.Sprintf("%s_%s.json", reportType, project.Name), "application/json")
	if err != nil {
		return fmt.Errorf("failed to upload report to S3: %w", err)
	}

	// Сохраняем запись о файле в БД
	_, err = pt.repo.CreateProjectFile(ctx, project.ID,
		fmt.Sprintf("%s_%s.json", reportType, project.Name),
		fmt.Sprintf("Отчет по чек-листу: %s", reportType),
		objectName, int64(len(reportJSON)), ".json", db.FileTypeFinalReport)
	if err != nil {
		return fmt.Errorf("failed to create project file record: %w", err)
	}

	log.Printf("Successfully saved checklist report %s to S3", objectName)

	// Устанавливаем статус ready после успешной обработки
	return pt.setProjectStatusReady(ctx, project.ID)
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

// generatePDFFromRemarks генерирует PDF отчет в стиле ГОСТ из замечаний
func (pt *ProjectProcessorTask) generatePDFFromRemarks(remarksResponse RemarksResponse) (*bytes.Buffer, error) {
	// Создаем новый PDF документ
	pdf := gofpdf.New("P", "mm", "A4", "")

	// Устанавливаем шрифт с поддержкой кириллицы
	pdf.AddUTF8Font("DejaVu", "", "DejaVuSans.ttf")
	pdf.AddUTF8Font("DejaVu", "B", "DejaVuSans-Bold.ttf")

	// Устанавливаем отступы (ГОСТ-подобные)
	pdf.SetMargins(30, 20, 10) // left, top, right

	// Добавляем первую страницу
	pdf.AddPage()

	// Титульная страница
	pdf.SetFont("DejaVu", "B", 18)
	pdf.Cell(0, 20, "ПАО «Газпром»")
	pdf.Ln(15)

	pdf.SetFont("DejaVu", "B", 16)
	pdf.Cell(0, 20, "Отчёт по результатам анализа замечаний")
	pdf.Ln(15)

	pdf.SetFont("DejaVu", "", 14)
	pdf.Cell(0, 20, fmt.Sprintf("Проект: %d", pt.projectID))
	pdf.Ln(15)

	pdf.SetFont("DejaVu", "", 12)
	pdf.Cell(0, 20, fmt.Sprintf("Дата: %s", time.Now().Format("02.01.2006")))
	pdf.Ln(30)

	// Оглавление
	pdf.AddPage()
	pdf.SetFont("DejaVu", "B", 16)
	pdf.Cell(0, 20, "СОДЕРЖАНИЕ")
	pdf.Ln(20)

	// Введение
	pdf.SetFont("DejaVu", "B", 14)
	pdf.Cell(0, 15, "ВВЕДЕНИЕ")
	pdf.Ln(15)

	pdf.SetFont("DejaVu", "", 12)
	introText := "Настоящий отчёт подготовлен на основании предоставленных данных. " +
		"Категории данных сформированы как главы, группы — как подразделы. " +
		"Для каждого подраздела приведены синтезированное описание и исходные замечания."

	// Разбиваем текст на строки для корректного отображения
	lines := pdf.SplitText(introText, 150)
	for _, line := range lines {
		pdf.Cell(0, 8, line)
		pdf.Ln(8)
	}
	pdf.Ln(10)

	// Основные разделы
	for section, items := range remarksResponse {
		// Заголовок раздела
		pdf.SetFont("DejaVu", "B", 14)
		pdf.Cell(0, 15, section)
		pdf.Ln(15)

		for _, item := range items {
			// Подзаголовок
			pdf.SetFont("DejaVu", "B", 12)
			pdf.Cell(0, 12, item.GroupName)
			pdf.Ln(12)

			// Синтезированное замечание
			if item.SynthesizedRemark != "" {
				pdf.SetFont("DejaVu", "B", 11)
				pdf.Cell(0, 10, "Краткая сводка:")
				pdf.Ln(10)

				pdf.SetFont("DejaVu", "", 11)
				synthLines := pdf.SplitText(item.SynthesizedRemark, 150)
				for _, line := range synthLines {
					pdf.Cell(0, 8, line)
					pdf.Ln(8)
				}
				pdf.Ln(5)
			}

			// Таблица с оригинальными замечаниями
			if len(item.OriginalDuplicates) > 0 {
				pdf.SetFont("DejaVu", "B", 11)
				pdf.Cell(0, 10, "Оригинальные замечания:")
				pdf.Ln(10)

				// Создаем таблицу
				colWidths := []float64{15, 135} // № | Замечание
				rowHeight := 8.0

				// Заголовок таблицы
				pdf.SetFont("DejaVu", "B", 10)
				pdf.SetFillColor(240, 240, 240)
				pdf.CellFormat(colWidths[0], rowHeight, "№", "1", 0, "C", true, 0, "")
				pdf.CellFormat(colWidths[1], rowHeight, "Замечание", "1", 0, "L", true, 0, "")
				pdf.Ln(-1)

				// Строки таблицы
				pdf.SetFont("DejaVu", "", 10)
				for i, remark := range item.OriginalDuplicates {
					// Номер
					pdf.CellFormat(colWidths[0], rowHeight, fmt.Sprintf("%d", i+1), "1", 0, "C", false, 0, "")

					// Замечание (разбиваем на строки если длинное)
					remarkLines := pdf.SplitText(remark, colWidths[1]-2)
					if len(remarkLines) == 1 {
						pdf.CellFormat(colWidths[1], rowHeight, remark, "1", 0, "L", false, 0, "")
						pdf.Ln(-1)
					} else {
						// Первая строка
						pdf.CellFormat(colWidths[1], rowHeight, remarkLines[0], "1", 0, "L", false, 0, "")
						pdf.Ln(-1)

						// Остальные строки
						for j := 1; j < len(remarkLines); j++ {
							pdf.CellFormat(colWidths[0], rowHeight, "", "1", 0, "C", false, 0, "")
							pdf.CellFormat(colWidths[1], rowHeight, remarkLines[j], "1", 0, "L", false, 0, "")
							pdf.Ln(-1)
						}
					}
				}
				pdf.Ln(10)
			}
		}
		pdf.Ln(10)
	}

	// Заключение
	pdf.AddPage()
	pdf.SetFont("DejaVu", "B", 14)
	pdf.Cell(0, 15, "ЗАКЛЮЧЕНИЕ")
	pdf.Ln(15)

	pdf.SetFont("DejaVu", "", 12)
	conclusionText := "Предложенные мероприятия направлены на снижение неопределённостей и повышение качества " +
		"прогнозов. Рекомендовано согласовать план доизучения и актуализировать модели по итогам " +
		"получения новых данных."

	conclusionLines := pdf.SplitText(conclusionText, 150)
	for _, line := range conclusionLines {
		pdf.Cell(0, 8, line)
		pdf.Ln(8)
	}

	// Сохраняем в буфер
	buffer := new(bytes.Buffer)
	err := pdf.Output(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buffer, nil
}
