package handler

import (
	"encoding/json"
	"errors"
	"evaluation/internal/postgres"
	"evaluation/internal/repository"
	"evaluation/internal/storage"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	m "evaluation/internal/models"
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
	pgClient *postgres.Client
	repo     *repository.Repository
	storage  storage.FileStorage
}

// New создает новый экземпляр хендлера
func New(pgClient *postgres.Client, repo *repository.Repository, fileStorage storage.FileStorage) *Handler {
	return &Handler{
		pgClient: pgClient,
		repo:     repo,
		storage:  fileStorage,
	}
}

// ========== HEALTH CHECK ==========

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Database  string `json:"database"`
}

// Health проверяет состояние сервиса
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем состояние базы данных
	dbStatus := "healthy"
	if err := h.pgClient.HealthCheck(r.Context()); err != nil {
		dbStatus = "unhealthy"
	}

	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().Format(time.RFC3339),
		Database:  dbStatus,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ========== FILE UPLOAD HANDLER ==========

// UploadAttach godoc
// @Summary Upload attach
// @Description Upload attach
// @ID uploadAttach
// @Accept  multipart/form-data
// @Produce  json
// @Param file formData file true "attach"
// @Param type query string true "type: excel or doc"
// @Success 200 {object} Response "ok"
// @Failure 400 {object} Error "bad request"
// @Failure 500 {object} Error "internal Server Error - Request is valid but operation failed at server side"
// @Router /attach [post]
func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		log.Println(m.StacktraceError(err))
		returnErrorJSON(w, m.ErrBadRequest400)
		return
	}
	defer file.Close()

	fileName := fileHeader.Filename
	log.Printf("Received file: %s", fileName)

	ext := strings.ToLower(filepath.Ext(fileName))
	log.Printf("File extension: '%s'", ext)

	if ext != ".xlsx" && ext != ".docx" {
		log.Printf("Invalid file extension: '%s'. Expected .xlsx or .docx", ext)
		returnErrorJSON(w, m.StacktraceError(errors.New("invalid file extension"), m.ErrServerError500))
		return
	}

	// Определяем MIME тип файла
	contentType := "application/octet-stream"
	if ext == ".xlsx" {
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	} else if ext == ".docx" {
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}

	// Загружаем файл в MinIO
	objectName, err := h.storage.UploadFile(r.Context(), file, fileName, contentType)
	if err != nil {
		log.Println(m.StacktraceError(err))
		returnErrorJSON(w, err)
		return
	}

	// Сохраняем информацию о файле в базе данных
	fileName, err = h.repo.SaveAttach(&m.Attach{
		Type:    r.URL.Query().Get("type"),
		File:    file,
		FileExt: ext,
	})
	if err != nil {
		log.Println(m.StacktraceError(err))
		returnErrorJSON(w, err)
		return
	}

	json.NewEncoder(w).Encode(&m.UploadAttachResponse{File: objectName})
}

// ========== PROJECT HANDLERS ==========

// HandleProjects обрабатывает запросы к /api/projects
func (h *Handler) HandleProjects(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListProjects(w, r)
	case http.MethodPost:
		h.CreateProject(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
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

	// Получаем список проектов
	projects, err := h.repo.ListProjects(r.Context())
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

// CreateProject godoc
// @Summary Create new project
// @Description Create a new project with name
// @ID createProject
// @Accept  json
// @Produce  json
// @Param project body CreateProjectRequest true "Project data"
// @Success 201 {object} db.Project "Project created successfully"
// @Failure 400 {object} Error "Bad request - invalid input data"
// @Failure 500 {object} Error "Internal server error"
// @Router /projects [post]
func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		returnErrorJSON(w, m.ErrBadRequest400)
		return
	}

	// Валидация входных данных
	if req.Name == "" {
		log.Println("Project name is required")
		returnErrorJSON(w, m.StacktraceError(errors.New("project name is required"), m.ErrBadRequest400))
		return
	}

	if len(req.Name) > 255 {
		log.Println("Project name too long")
		returnErrorJSON(w, m.StacktraceError(errors.New("project name too long (max 255 characters)"), m.ErrBadRequest400))
		return
	}

	// Создаем проект (in_progress всегда false по умолчанию)
	project, err := h.repo.CreateProject(r.Context(), req.Name, false)
	if err != nil {
		log.Printf("Failed to create project: %v", err)
		returnErrorJSON(w, m.ErrServerError500)
		return
	}

	// Возвращаем созданный проект
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(project)
}

// CreateProjectRequest структура запроса для создания проекта
type CreateProjectRequest struct {
	Name string `json:"name" validate:"required,max=255"`
}
