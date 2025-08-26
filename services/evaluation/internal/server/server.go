package server

import (
	"context"
	"fmt"
	"net/http"

	"evaluation/internal/config"
	"evaluation/internal/handler"
	"evaluation/internal/services"
	"evaluation/internal/tasks"

	_ "evaluation/internal/handler/docs"

	"github.com/gorilla/mux"
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

	// Создаем роутер с gorilla/mux
	r := mux.NewRouter()

	// Добавляем CORS middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Разрешаем все origin для разработки
			origin := r.Header.Get("Origin")

			// Разрешаем конкретные origin'ы (настройте под свои нужды)
			allowedOrigins := map[string]bool{
				"http://127.0.0.1:5001": true,
				"http://localhost:5001": true,
				"http://127.0.0.1:5000": true,
				"http://localhost:5000": true,
			}

			if allowedOrigins[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "3600")

			// Обрабатываем preflight OPTIONS запросы
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Health check endpoint
	r.HandleFunc("/health", handler.Health).Methods("GET")

	// API endpoints
	r.HandleFunc("/api/projects", handler.HandleProjects).Methods("GET", "POST", "OPTIONS")

	// Специфичные пути для проектов с поддержкой параметров
	r.HandleFunc("/api/projects/{id:[0-9]+}", handler.HandleProject).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/projects/{id:[0-9]+}/files", handler.HandleProjectFiles).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/projects/{id:[0-9]+}/final_report", handler.HandleGenerateFinalReport).Methods("POST", "OPTIONS")

	// GET ручки для получения результатов обработки
	r.HandleFunc("/api/projects/{id:[0-9]+}/checklist", handler.HandleGetChecklist).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/projects/{id:[0-9]+}/remarks_clustered", handler.HandleGetRemarksClustered).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/projects/{id:[0-9]+}/final_report", handler.HandleGetFinalReport).Methods("GET", "OPTIONS")

	// Swagger docs
	r.PathPrefix("/api/docs/").Handler(httpSwagger.WrapHandler)

	// Создаем HTTP сервер
	httpServer := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
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
