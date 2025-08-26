package server

import (
	"context"
	"fmt"
	"net/http"

	"evaluation/internal/config"
	"evaluation/internal/handler"
	"evaluation/internal/postgres"
	"evaluation/internal/repository"
	"evaluation/internal/storage"
	"evaluation/internal/tasks"

	_ "evaluation/internal/handler/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

type Server struct {
	httpServer  *http.Server
	config      *config.Config
	pgClient    *postgres.Client
	repo        *repository.Repository
	storage     storage.FileStorage
	taskManager tasks.TaskManager
}

func New(cfg *config.Config, pgClient *postgres.Client, repo *repository.Repository, fileStorage storage.FileStorage, taskManager tasks.TaskManager) *Server {
	// Создаем единый хендлер
	handler := handler.New(pgClient, repo, fileStorage, taskManager)

	// Настраиваем роутинг
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", handler.Health)

	// API endpoints
	mux.HandleFunc("/api/attach", handler.UploadFile)
	mux.HandleFunc("/api/projects", handler.HandleProjects)
	mux.HandleFunc("/api/projects/{id}", handler.HandleProject)
	// mux.HandleFunc("/api/project-files", handler.HandleProjectFiles)
	// mux.HandleFunc("/api/project-files/", handler.HandleProjectFile)

	// mux.HandleFunc("/api/attach", handler.UploadFile)
	// mux.HandleFunc("/api/file", handler.SendFile)
	// mux.HandleFunc("/api/projects/{project_id}/remarks", handler.SendProjectRemarks)
	mux.HandleFunc("/api/projects/{id}/files", handler.HandleProjectFiles)
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
		httpServer: httpServer,
		config:     cfg,
		pgClient:   pgClient,
		repo:       repo,
		storage:    fileStorage,
	}
}

func (s *Server) Start() error {
	fmt.Printf("Server starting on port %s\n", s.config.Server.Port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
