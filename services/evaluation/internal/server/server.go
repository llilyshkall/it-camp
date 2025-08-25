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

	_ "evaluation/internal/handler/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

type Server struct {
	httpServer *http.Server
	config     *config.Config
	pgClient   *postgres.Client
	repo       *repository.Repository
	storage    storage.FileStorage
}

// func loggingAndCORSHeadersMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		log.Println(r.RequestURI, r.Method)
// 		for header := range Headers {
// 			w.Header().Set(header, Headers[header])
// 		}

// 		next.ServeHTTP(w, r)
// 	})
// }

// var Headers = map[string]string{
// 	//"Access-Control-Allow-Origin":      "http://127.0.0.1:8001",
// 	"Access-Control-Allow-Credentials": "true",
// 	"Access-Control-Allow-Headers":     "Origin, Content-Type, accept, csrf",
// 	"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, OPTIONS",
// 	"Content-Type":                     "application/json",
// }

func New(cfg *config.Config, pgClient *postgres.Client, repo *repository.Repository, fileStorage storage.FileStorage) *Server {
	// Создаем единый хендлер
	handler := handler.New(pgClient, repo, fileStorage)

	// Настраиваем роутинг
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", handler.Health)

	// API endpoints
	mux.HandleFunc("/api/projects", handler.HandleProjects)
	mux.HandleFunc("/api/projects/", handler.HandleProject)
	//mux.HandleFunc("/api/remarks", handler.HandleRemarks)
	//mux.HandleFunc("/api/remarks/", handler.HandleRemark)
	mux.HandleFunc("/api/project-files", handler.HandleProjectFiles)
	mux.HandleFunc("/api/project-files/", handler.HandleProjectFile)

	mux.HandleFunc("/api/attach", handler.UploadFile)
	//mux.HandleFunc("/api/docs/", handler.UploadFile)
	mux.HandleFunc("/api/docs/", httpSwagger.WrapHandler)
	//mux.Use(loggingAndCORSHeadersMiddleware)
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
