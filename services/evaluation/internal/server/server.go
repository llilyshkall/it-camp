package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"evaluation/internal/config"
	"evaluation/internal/handler"
	"evaluation/internal/services"
	"evaluation/internal/tasks"

	_ "evaluation/internal/handler/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

type Server struct {
	httpServer     *http.Server
	config         *config.Config
	projectService services.ProjectService
	fileService    services.FileService
	healthService  services.HealthService
	taskManager    tasks.TaskManager
}

func New(cfg *config.Config, projectService services.ProjectService, fileService services.FileService, healthService services.HealthService, taskManager tasks.TaskManager) *Server {
	// Создаем единый хендлер
	handler := handler.New(projectService, fileService, healthService, taskManager)

	// Настраиваем роутинг
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", handler.Health)

	// API endpoints
	mux.HandleFunc("/api/projects", handler.HandleProjects)

	// Для работы с конкретным проектом используем более специфичные пути
	mux.HandleFunc("/api/projects/", func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем путь и определяем, какой хендлер использовать
		path := r.URL.Path

		// Если путь заканчивается на /files, используем HandleProjectFiles
		if strings.HasSuffix(path, "/files") {
			handler.HandleProjectFiles(w, r)
			return
		}

		// Если путь заканчивается на /final_report, используем HandleGenerateFinalReport
		if strings.HasSuffix(path, "/final_report") {
			handler.HandleGenerateFinalReport(w, r)
			return
		}

		// Иначе используем HandleProject для получения/обновления проекта
		handler.HandleProject(w, r)
	})

	mux.HandleFunc("/api/docs/", httpSwagger.WrapHandler)

	// Создаем HTTP сервер
	httpServer := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &Server{
		httpServer:     httpServer,
		config:         cfg,
		projectService: projectService,
		fileService:    fileService,
		healthService:  healthService,
		taskManager:    taskManager,
	}
}

func (s *Server) Start() error {
	fmt.Printf("Server starting on port %s\n", s.config.Server.Port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
