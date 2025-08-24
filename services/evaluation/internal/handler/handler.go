package handler

import (
	"encoding/json"
	"evaluation/internal/postgres"
	"evaluation/internal/repository"
	"io"
	"log"
	"net/http"
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

// @host 127.0.0.1:8080
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

// Handler объединяет все HTTP хендлеры
type Handler struct {
	pgClient *postgres.Client
	repo     *repository.Repository
}

// New создает новый экземпляр хендлера
func New(pgClient *postgres.Client, repo *repository.Repository) *Handler {
	return &Handler{
		pgClient: pgClient,
		repo:     repo,
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

	w.Header().Set("Content-Type", "application/json")
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
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	file, _, err := r.FormFile("file")
	if err != nil {
		log.Println(m.StacktraceError(err))
		//returnErrorJSON(w, e.ErrBadRequest400)
		//return
		//log.Println(err)
		http.Error(w, "Failed to get project file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	fileName, err := h.repo.SaveAttach(&m.Attach{
		Type: r.URL.Query().Get("type"),
		File: file,
	})
	if err != nil {
		//log.Println(err)
		log.Println(m.StacktraceError(err))
		//returnErrorJSON(w, err)
		http.Error(w, "Failed to get project file", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(&m.UploadAttachResponse{File: fileName})
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

// ========== REMARK HANDLERS ==========

// HandleRemarks обрабатывает запросы к /api/remarks
// func (h *Handler) HandleRemarks(w http.ResponseWriter, r *http.Request) {
// 	switch r.Method {
// 	case http.MethodGet:
// 		h.ListRemarks(w, r)
// 	case http.MethodPost:
// 		h.CreateRemark(w, r)
// 	default:
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 	}
// }

// // HandleRemark обрабатывает запросы к /api/remarks/{id}
// func (h *Handler) HandleRemark(w http.ResponseWriter, r *http.Request) {
// 	// Извлекаем ID из URL
// 	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
// 	if len(pathParts) < 3 {
// 		http.Error(w, "Invalid remark ID", http.StatusBadRequest)
// 		return
// 	}

// 	id, err := strconv.ParseInt(pathParts[2], 10, 32)
// 	if err != nil {
// 		http.Error(w, "Invalid remark ID", http.StatusBadRequest)
// 		return
// 	}

// 	switch r.Method {
// 	case http.MethodGet:
// 		h.GetRemark(w, r, int32(id))
// 	case http.MethodPut:
// 		h.UpdateRemark(w, r, int32(id))
// 	case http.MethodDelete:
// 		h.DeleteRemark(w, r, int32(id))
// 	default:
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 	}
// }

// func (h *Handler) ListRemarks(w http.ResponseWriter, r *http.Request) {
// 	// Получаем project_id из query параметров
// 	projectIDStr := r.URL.Query().Get("project_id")
// 	if projectIDStr == "" {
// 		http.Error(w, "project_id is required", http.StatusBadRequest)
// 		return
// 	}

// 	projectID, err := strconv.ParseInt(projectIDStr, 10, 32)
// 	if err != nil {
// 		http.Error(w, "Invalid project_id", http.StatusBadRequest)
// 		return
// 	}

// 	remarks, err := h.repo.ListRemarksByProject(r.Context(), int32(projectID))
// 	if err != nil {
// 		http.Error(w, "Failed to list remarks", http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(remarks)
// }

// func (h *Handler) GetRemark(w http.ResponseWriter, r *http.Request, id int32) {
// 	remark, err := h.repo.GetRemark(r.Context(), id)
// 	if err != nil {
// 		http.Error(w, "Remark not found", http.StatusNotFound)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(remark)
// }

// func (h *Handler) CreateRemark(w http.ResponseWriter, r *http.Request) {
// 	var req struct {
// 		ProjectID  int32  `json:"project_id"`
// 		Direction  string `json:"direction"`
// 		Section    string `json:"section"`
// 		Subsection string `json:"subsection"`
// 		Content    string `json:"content"`
// 	}

// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	remark, err := h.repo.CreateRemark(r.Context(), req.ProjectID, req.Direction, req.Section, req.Subsection, req.Content)
// 	if err != nil {
// 		http.Error(w, "Failed to create remark", http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(remark)
// }

// func (h *Handler) UpdateRemark(w http.ResponseWriter, r *http.Request, id int32) {
// 	var req struct {
// 		Direction  string `json:"direction"`
// 		Section    string `json:"section"`
// 		Subsection string `json:"subsection"`
// 		Content    string `json:"content"`
// 	}

// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	remark, err := h.repo.UpdateRemark(r.Context(), id, req.Direction, req.Section, req.Subsection, req.Content)
// 	if err != nil {
// 		http.Error(w, "Failed to update remark", http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(remark)
// }

// func (h *Handler) DeleteRemark(w http.ResponseWriter, r *http.Request, id int32) {
// 	if err := h.repo.DeleteRemark(r.Context(), id); err != nil {
// 		http.Error(w, "Failed to delete remark", http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusNoContent)
// }

// ========== PROJECT FILE HANDLERS ==========

// HandleProjectFiles обрабатывает запросы к /api/project-files
func (h *Handler) HandleProjectFiles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListProjectFiles(w, r)
	case http.MethodPost:
		h.CreateProjectFile(w, r)
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

func (h *Handler) CreateProjectFile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProjectID    int32  `json:"project_id"`
		Filename     string `json:"filename"`
		OriginalName string `json:"original_name"`
		FilePath     string `json:"file_path"`
		FileSize     int64  `json:"file_size"`
		MimeType     string `json:"mime_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	file, err := h.repo.CreateProjectFile(r.Context(), req.ProjectID, req.Filename, req.OriginalName, req.FilePath, req.FileSize, req.MimeType)
	if err != nil {
		http.Error(w, "Failed to create project file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(file)
}

func (h *Handler) DeleteProjectFile(w http.ResponseWriter, r *http.Request, id int32) {
	if err := h.repo.DeleteProjectFile(r.Context(), id); err != nil {
		http.Error(w, "Failed to delete project file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// type File interface {
// 	io.Reader
// 	io.ReaderAt
// 	io.Seeker
// 	io.Closer
// }

// type Attach struct {
// 	Type string
// 	File File
// }

// type UploadAttachResponse struct {
// 	File string `json:"file"`
// }
