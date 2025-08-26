package handler

import (
	"bytes"
	"encoding/json"
	"evaluation/internal/services"
	"evaluation/internal/tasks"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	m "evaluation/internal/models"

	"github.com/gorilla/mux"
)

// @title API
// @version 1.0
// @description ... back server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host 127.0.0.1:8081
// @BasePath  /api

type File interface {
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Closer
}

type Attach struct {
	Type    string
	File    File
	FileExt string
}

type UploadAttachResponse struct {
	File string `json:"file"`
}

type Error struct {
	Error interface{} `json:"error,omitempty"`
}

type Response struct {
	Body interface{} `json:"body,omitempty"`
}

func returnErrorJSON(w http.ResponseWriter, err error) {
	errCode, errText := m.CheckError(err)
	w.WriteHeader(errCode)
	json.NewEncoder(w).Encode(&m.Error{Error: errText})
}

// Handler объединяет все HTTP хендлеры
type Handler struct {
	projectService services.ProjectService
	fileService    services.FileService
	healthService  services.HealthService
	taskManager    tasks.TaskManager
}

// New создает новый экземпляр хендлера
func New(projectService services.ProjectService, fileService services.FileService, healthService services.HealthService, taskManager tasks.TaskManager) *Handler {
	return &Handler{
		projectService: projectService,
		fileService:    fileService,
		healthService:  healthService,
		taskManager:    taskManager,
	}
}

// ========== HEALTH CHECK ==========

// Health проверяет состояние сервиса
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем состояние сервиса через сервис
	response, err := h.healthService.CheckHealth(r.Context())
	if err != nil {
		log.Printf("Health check failed: %v", err)
		returnErrorJSON(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// SendFile godoc
// @Summary Send file
// @Description Send file
// @ID sendFile
// @Accept  json
// @Produce  octet-stream
// @Success 200 {file} file "File attachment"
// @Failure 400 {object} Error "bad request"
// @Failure 500 {object} Error "internal Server Error - Request is valid but operation failed at server side"
// @Router /file [get]
func (h *Handler) SendFile(w http.ResponseWriter, r *http.Request) {
	// Разрешить CORS для всех источников (для разработки)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Обработка preflight-запроса
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	dir, err := os.Getwd()
	if err != nil {
		returnErrorJSON(w, m.StacktraceError(errors.New("invalid file location"), m.ErrServerError500))
	}
	filename := "aed85cd5-53d7-4eb6-a106-4deee07ed2a1.xlsx"
	filePath := dir + "/" + filename
	//filepath = ""
	log.Println(filePath)
	// // Проверяем существование файла
	// if _, err := os.Stat(filePath); os.IsNotExist(err) {
	// 	http.Error(w, "File not found", http.StatusNotFound)
	// 	return
	// }

	// // Получаем имя файла из пути
	// _, fileName := filepath.Split(filePath)

	//json.NewEncoder(w).Encode(&m.Response{Body: filePath})
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Expires", "0")
	http.ServeFile(w, r, filePath)
}

// SendProjectRemarks godoc
// @Summary Send Project Remarks
// @Description Get remarks for specific project and forward to external service
// @ID sendProjectRemarks
// @Accept  json
// @Produce json
// @Param project_id path string true "Project ID"
// @Success 200 {object} Response "Success response"
// @Failure 400 {object} Error "bad request"
// @Failure 404 {object} Error "project not found"
// @Failure 500 {object} Error "internal server error"
// @Router /projects/{project_id}/remarks [post]
func (h *Handler) SendProjectRemarks(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	projectID, err := strconv.ParseInt(pathParts[2], 10, 32)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// Extract project_id from path
	// vars := mux.Vars(r)
	// projectID := vars["project_id"]
	if projectID == 0 {
		returnErrorJSON(w, m.StacktraceError(errors.New("project_id is required"), m.ErrBadRequest400))
		return
	}

	// Get project data from database
	// projectRemarks, err := h.Repository.GetProjectByID(projectID)
	// if err != nil {
	// 	if errors.Is(err, sql.ErrNoRows) {
	// 		returnErrorJSON(w, m.StacktraceError(fmt.Errorf("project %s not found", projectID), m.ErrNotFound404))
	// 	} else {
	// 		returnErrorJSON(w, m.StacktraceError(err, m.ErrServerError500))
	// 	}
	// 	return
	// }
	projectRemarks := map[string]interface{}{
		"None": []string{
			"Показать, как прогнозируется распространение водонасыщенных линз в геологической модели и их влияние на НГЗ",
			"ТРебуется более криичное рассмотрение геологии в районе грабена",
		},
		"development": []string{
			"Система ППД необоснована",
		},
		"geological": []string{
			"Неочевидно влияние переходной зоны",
		},
		"hydrodynamic_integrated": []string{
			"моделей нет почему",
		},
		"petrophysical": []string{
			"Запланировать исследования керна по скважине 8306 на расклинивающий эффект",
			"Привести данные лабораторных исследований параметра пористости на графике Рп-Кп по скважине 8306 в пластовых условиях",
			"Провести сравнительный анализ результатов комплекса ГИС \"новых\" и \"исторических\" скважин.",
		},
		"reassessment": []string{
			"Привести на отдельном слайде сравнение плановых и фактических показателей по скважинам ОПР. Показать плановые и фактические показатели Кпрод на исторических скважинах.",
			"При обосновании контактов по блокам на планшетах показать фактические притоки по испытаниям в колонне или открытом стволе",
		},
		"seismogeological": []string{
			"Провести ретроспективный анализ прогнозной способности куба АИ эффективных толщин по циклитам. Сравнить плановые показатели песчанистости из ГМ 2022 г и фактические показатели, полученные в скважинах ОПР. Показать отклонения в цифрах.",
		},
		// Вложенная карта для ключей
		"keys": map[string]string{
			"reassessment":            "Программа доизучения (ГРР и ОПР)",
			"seismogeological":        "Сейсмогеологическая модель",
			"petrophysical":           "Петрофизическая модель",
			"geological":              "Геологическая модель",
			"development":             "Разработка и прогноз технологических показателей добычи",
			"hydrodynamic_integrated": "Гидродинамическая и интегрированная модели",
		},
	}
	// Prepare request to external service
	externalURL := "http://127.0.0.1:8083/remarks"
	jsonData, err := json.Marshal(projectRemarks)
	if err != nil {
		returnErrorJSON(w, m.StacktraceError(err, m.ErrServerError500))
		return
	}

	// Send request to external service
	resp, err := http.Post(externalURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		returnErrorJSON(w, m.StacktraceError(fmt.Errorf("failed to send remarks to external service: %v", err), m.ErrServerError500))
		return
	}
	defer resp.Body.Close()

	// Check external service response
	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		returnErrorJSON(w, m.StacktraceError(
			fmt.Errorf("external service returned status %d: %s", resp.StatusCode, string(body)),
			m.ErrServerError500,
		))
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Remarks for project %s processed successfully", projectID),
	})
}

// ========== PROJECT HANDLERS ==========

// HandleProjects обрабатывает запросы к /api/projects
func (h *Handler) HandleProjects(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.ListProjects(w, r)
	case http.MethodPost:
		h.CreateProject(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleProject обрабатывает запросы к /api/projects/{id}
func (h *Handler) HandleProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.GetProject(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleProjectFiles обрабатывает запросы к /api/projects/{id}/files
func (h *Handler) HandleProjectFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	switch r.Method {
	case http.MethodPost:
		h.UploadProjectFile(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleGenerateFinalReport обрабатывает запросы к /api/projects/{id}/final_report
func (h *Handler) HandleGenerateFinalReport(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.GenerateFinalReport(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// UploadProjectFile godoc
// @Summary Upload file to project
// @Description Upload a file to a specific project with specified type
// @ID uploadProjectFile
// @Accept  multipart/form-data
// @Produce  json
// @Param id path int true "Project ID"
// @Param file formData file true "File to upload"
// @Param type query string true "Type of file: documentation, remarks, checklist, final_report, remarks_clustered"
// @Success 202 {object} db.ProjectFile "File uploaded successfully"
// @Failure 400 {object} Error "Bad request - invalid input data"
// @Failure 404 {object} Error "Project not found"
// @Failure 500 {object} Error "Internal server error"
// @Router /projects/{id}/files [post]
func (h *Handler) UploadProjectFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем ID проекта из URL с помощью gorilla/mux
	vars := mux.Vars(r)
	projectIDStr, ok := vars["id"]
	if !ok {
		log.Println("Project ID not found in URL")
		returnErrorJSON(w, m.ErrBadRequest400)
		return
	}

	projectID, err := strconv.ParseInt(projectIDStr, 10, 32)
	if err != nil {
		log.Printf("Invalid project ID format: %v", err)
		returnErrorJSON(w, m.ErrBadRequest400)
		return
	}

	// Получаем тип файла из query параметров
	fileTypeStr := r.URL.Query().Get("type")

	// Ограничиваем размер файла
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20) // 50MB

	// Получаем файл из формы
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		log.Printf("Failed to get file: %v", err)
		returnErrorJSON(w, m.ErrBadRequest400)
		return
	}
	defer file.Close()

	fileName := fileHeader.Filename
	fileSize := fileHeader.Size
	log.Printf("Received file: %s, size: %d bytes", fileName, fileSize)

	// Используем сервис для загрузки файла в проект
	projectFile, err := h.fileService.UploadProjectFile(r.Context(), int32(projectID), file, fileName, fileTypeStr, fileSize)
	if err != nil {
		log.Printf("Failed to upload project file: %v", err)
		returnErrorJSON(w, err)
		return
	}

	// Возвращаем информацию о загруженном файле
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(projectFile)
}

// ListProjects godoc
// @Summary Get all projects
// @Description Get list of all projects
// @ID listProjects
// @Accept  json
// @Produce  json
// @Success 200 {array} db.Project "List of projects"
// @Failure 500 {object} Error "Internal server error"
// @Router /projects [get]
func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем список проектов через сервис
	projects, err := h.projectService.ListProjects(r.Context())
	if err != nil {
		log.Printf("Failed to list projects: %v", err)
		returnErrorJSON(w, m.ErrServerError500)
		return
	}

	// Возвращаем список проектов
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(projects)
}

// GetProject godoc
// @Summary Get project by ID
// @Description Get specific project by its ID
// @ID getProject
// @Accept  json
// @Produce  json
// @Param id path int true "Project ID"
// @Success 200 {object} db.Project "Project found"
// @Failure 400 {object} Error "Bad request - invalid project ID"
// @Failure 404 {object} Error "Project not found"
// @Failure 500 {object} Error "Internal server error"
// @Router /projects/{id} [get]
func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем ID из URL с помощью gorilla/mux
	vars := mux.Vars(r)
	projectIDStr, ok := vars["id"]
	if !ok {
		log.Println("Project ID not found in URL")
		returnErrorJSON(w, m.ErrBadRequest400)
		return
	}

	id, err := strconv.ParseInt(projectIDStr, 10, 32)
	if err != nil {
		log.Printf("Invalid project ID format: %v", err)
		returnErrorJSON(w, m.ErrBadRequest400)
		return
	}

	// Получаем проект через сервис
	project, err := h.projectService.GetProject(r.Context(), int32(id))
	if err != nil {
		log.Printf("Failed to get project %d: %v", id, err)
		returnErrorJSON(w, m.ErrNotFound404)
		return
	}

	// Возвращаем проект
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(project)
}

// CreateProject godoc
// @Summary Create new project
// @Description Create a new project with name
// @ID createProject
// @Accept  json
// @Produce  json
// @Param project body models.CreateProjectRequest true "Project data"
// @Success 201 {object} db.Project "Project created successfully"
// @Failure 400 {object} Error "Bad request - invalid input data"
// @Failure 500 {object} Error "Internal server error"
// @Router /projects [post]
func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req m.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		returnErrorJSON(w, m.ErrBadRequest400)
		return
	}

	// Создаем проект через сервис
	project, err := h.projectService.CreateProject(r.Context(), req.Name)
	if err != nil {
		log.Printf("Failed to create project: %v", err)
		returnErrorJSON(w, err)
		return
	}

	// Возвращаем созданный проект
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(project)
}

// GenerateFinalReport godoc
// @Summary Generate final report for project
// @Description Generate final report for a specific project
// @ID generateFinalReport
// @Accept  json
// @Produce  json
// @Param id path int true "Project ID"
// @Success 202 {object} Response "Final report generation started"
// @Failure 400 {object} Error "Bad request - invalid project ID"
// @Failure 404 {object} Error "Project not found"
// @Failure 409 {object} Error "Project is already being processed"
// @Failure 500 {object} Error "Internal server error"
// @Router /projects/{id}/final_report [post]
func (h *Handler) GenerateFinalReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем ID проекта из URL с помощью gorilla/mux
	vars := mux.Vars(r)
	projectIDStr, ok := vars["id"]
	if !ok {
		log.Println("Project ID not found in URL")
		returnErrorJSON(w, m.ErrBadRequest400)
		return
	}

	projectID, err := strconv.ParseInt(projectIDStr, 10, 32)
	if err != nil {
		log.Printf("Invalid project ID format: %v", err)
		returnErrorJSON(w, m.ErrBadRequest400)
		return
	}

	// Используем сервис для генерации финального отчета
	err = h.fileService.GenerateFinalReport(r.Context(), int32(projectID))
	if err != nil {
		log.Printf("Failed to generate final report for project %d: %v", projectID, err)
		returnErrorJSON(w, err)
		return
	}

	// Возвращаем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(&Response{
		Body: map[string]interface{}{
			"message":    "Final report generation started",
			"project_id": projectID,
		},
	})
}

// HandleGetChecklist обрабатывает запросы к /api/projects/{id}/checklist
func (h *Handler) HandleGetChecklist(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetChecklist(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleGetRemarksClustered обрабатывает запросы к /api/projects/{id}/remarks_clustered
func (h *Handler) HandleGetRemarksClustered(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetRemarksClustered(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleGetFinalReport обрабатывает запросы к /api/projects/{id}/final_report
func (h *Handler) HandleGetFinalReport(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetFinalReport(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GetChecklist godoc
// @Summary Get project checklist
// @Description Get checklist result for a specific project
// @ID getChecklist
// @Accept  json
// @Produce  json
// @Param id path int true "Project ID"
// @Success 200 {object} Response "Checklist result"
// @Failure 400 {object} Error "Bad request - invalid project ID"
// @Failure 404 {object} Error "Project not found"
// @Failure 409 {object} Error "Checklist is still being generated"
// @Failure 500 {object} Error "Internal server error"
// @Router /projects/{id}/checklist [get]
func (h *Handler) GetChecklist(w http.ResponseWriter, r *http.Request) {
	// Извлекаем ID проекта из URL с помощью gorilla/mux
	vars := mux.Vars(r)
	projectIDStr, ok := vars["id"]
	if !ok {
		log.Println("Project ID not found in URL")
		returnErrorJSON(w, m.ErrBadRequest400)
		return
	}

	projectID, err := strconv.ParseInt(projectIDStr, 10, 32)
	if err != nil {
		log.Printf("Invalid project ID format: %v", err)
		returnErrorJSON(w, m.ErrBadRequest400)
		return
	}

	// Используем сервис для получения чеклиста
	result, err := h.fileService.GetChecklist(r.Context(), int32(projectID))
	if err != nil {
		log.Printf("Failed to get checklist for project %d: %v", projectID, err)
		returnErrorJSON(w, err)
		return
	}

	// Возвращаем результат
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&Response{
		Body: result,
	})
}

// GetRemarksClustered godoc
// @Summary Get clustered remarks for project
// @Description Get clustered remarks result for a specific project
// @ID getRemarksClustered
// @Accept  json
// @Produce  json
// @Param id path int true "Project ID"
// @Success 200 {object} Response "Clustered remarks result"
// @Failure 400 {object} Error "Bad request - invalid project ID"
// @Failure 404 {object} Error "Project not found"
// @Failure 409 {object} Error "Remarks are still being processed"
// @Failure 500 {object} Error "Internal server error"
// @Router /projects/{id}/remarks_clustered [get]
func (h *Handler) GetRemarksClustered(w http.ResponseWriter, r *http.Request) {
	// Извлекаем ID проекта из URL с помощью gorilla/mux
	vars := mux.Vars(r)
	projectIDStr, ok := vars["id"]
	if !ok {
		log.Println("Project ID not found in URL")
		returnErrorJSON(w, m.ErrBadRequest400)
		return
	}

	projectID, err := strconv.ParseInt(projectIDStr, 10, 32)
	if err != nil {
		log.Printf("Invalid project ID format: %v", err)
		returnErrorJSON(w, m.ErrBadRequest400)
		return
	}

	// Используем сервис для получения кластеризированных замечаний
	result, err := h.fileService.GetRemarksClustered(r.Context(), int32(projectID))
	if err != nil {
		log.Printf("Failed to get clustered remarks for project %d: %v", projectID, err)
		returnErrorJSON(w, err)
		return
	}

	// Возвращаем результат
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&Response{
		Body: result,
	})
}

// GetFinalReport godoc
// @Summary Get final report for project
// @Description Get final report result for a specific project
// @ID getFinalReport
// @Accept  json
// @Produce  json
// @Param id path int true "Project ID"
// @Success 200 {object} Response "Final report result"
// @Failure 400 {object} Error "Bad request - invalid project ID"
// @Failure 404 {object} Error "Project not found"
// @Failure 409 {object} Error "Final report is still being generated"
// @Failure 500 {object} Error "Internal server error"
// @Router /projects/{id}/final_report [get]
func (h *Handler) GetFinalReport(w http.ResponseWriter, r *http.Request) {
	// Извлекаем ID проекта из URL с помощью gorilla/mux
	vars := mux.Vars(r)
	projectIDStr, ok := vars["id"]
	if !ok {
		log.Println("Project ID not found in URL")
		returnErrorJSON(w, m.ErrBadRequest400)
		return
	}

	projectID, err := strconv.ParseInt(projectIDStr, 10, 32)
	if err != nil {
		log.Printf("Invalid project ID format: %v", err)
		returnErrorJSON(w, m.ErrBadRequest400)
		return
	}

	// Используем сервис для получения финального отчета
	result, err := h.fileService.GetFinalReport(r.Context(), int32(projectID))
	if err != nil {
		log.Printf("Failed to get final report for project %d: %v", projectID, err)
		returnErrorJSON(w, err)
		return
	}

	// Возвращаем результат
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&Response{
		Body: result,
	})
}
