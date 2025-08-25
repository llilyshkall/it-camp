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
	"os"
	"path/filepath"
	"strconv"
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

	//w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ========== PROJECT HANDLERS ==========

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
	// Разрешить CORS для всех источников (для разработки)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Обработка preflight-запроса
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

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

	json.NewEncoder(w).Encode(&m.Response{Body: filePath})
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Expires", "0")
	http.ServeFile(w, r, filePath)
}

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

// HandleProject обрабатывает запросы к /api/projects/{id}
func (h *Handler) HandleProject(w http.ResponseWriter, r *http.Request) {
	// Извлекаем ID из URL
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(pathParts[2], 10, 32)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.GetProject(w, r, int32(id))
	case http.MethodPut:
		h.UpdateProject(w, r, int32(id))
	case http.MethodDelete:
		h.DeleteProject(w, r, int32(id))
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.repo.ListProjects(r.Context())
	if err != nil {
		http.Error(w, "Failed to list projects", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(projects)
}

func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request, id int32) {
	project, err := h.repo.GetProject(r.Context(), id)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(project)
}

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name       string `json:"name"`
		InProgress bool   `json:"in_progress"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	project, err := h.repo.CreateProject(r.Context(), req.Name, req.InProgress)
	if err != nil {
		http.Error(w, "Failed to create project", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(project)
}

func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request, id int32) {
	var req struct {
		Name       string `json:"name"`
		InProgress bool   `json:"in_progress"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	project, err := h.repo.UpdateProject(r.Context(), id, req.Name, req.InProgress)
	if err != nil {
		http.Error(w, "Failed to update project", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(project)
}

func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request, id int32) {
	if err := h.repo.DeleteProject(r.Context(), id); err != nil {
		http.Error(w, "Failed to delete project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ========== PROJECT FILE HANDLERS ==========

// HandleProjectFiles обрабатывает запросы к /api/project-files
func (h *Handler) HandleProjectFiles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListProjectFiles(w, r)
	// case http.MethodPost:
	// 	h.CreateProjectFile(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleProjectFile обрабатывает запросы к /api/project-files/{id}
func (h *Handler) HandleProjectFile(w http.ResponseWriter, r *http.Request) {
	// Извлекаем ID из URL
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid project file ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(pathParts[2], 10, 32)
	if err != nil {
		http.Error(w, "Invalid project file ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.GetProjectFile(w, r, int32(id))
	case http.MethodDelete:
		h.DeleteProjectFile(w, r, int32(id))
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) ListProjectFiles(w http.ResponseWriter, r *http.Request) {
	// Получаем project_id из query параметров
	projectIDStr := r.URL.Query().Get("project_id")
	if projectIDStr == "" {
		http.Error(w, "project_id is required", http.StatusBadRequest)
		return
	}

	projectID, err := strconv.ParseInt(projectIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid project_id", http.StatusBadRequest)
		return
	}

	files, err := h.repo.ListProjectFiles(r.Context(), int32(projectID))
	if err != nil {
		http.Error(w, "Failed to list project files", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(files)
}

func (h *Handler) GetProjectFile(w http.ResponseWriter, r *http.Request, id int32) {
	file, err := h.repo.GetProjectFile(r.Context(), id)
	if err != nil {
		http.Error(w, "Project file not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(file)
}

// func (h *Handler) CreateProjectFile(w http.ResponseWriter, r *http.Request) {
// 	var req struct {
// 		ProjectID    int32  `json:"project_id"`
// 		Filename     string `json:"filename"`
// 		OriginalName string `json:"original_name"`
// 		FilePath     string `json:"file_path"`
// 		FileSize     int64  `json:"file_size"`
// 		MimeType     string `json:"mime_type"`
// 	}

// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	file, err := h.repo.CreateProjectFile(r.Context(), req.ProjectID, req.Filename, req.OriginalName, req.FilePath, req.FileSize, req.MimeType)
// 	if err != nil {
// 		http.Error(w, "Failed to create project file", http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(file)
// }

func (h *Handler) DeleteProjectFile(w http.ResponseWriter, r *http.Request, id int32) {
	if err := h.repo.DeleteProjectFile(r.Context(), id); err != nil {
		http.Error(w, "Failed to delete project file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
