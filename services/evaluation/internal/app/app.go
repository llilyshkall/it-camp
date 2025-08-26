package app

import (
	"context"
	"evaluation/internal/config"
	"evaluation/internal/postgres"
	"evaluation/internal/repository"
	"evaluation/internal/server"
	"evaluation/internal/storage"
	"evaluation/internal/tasks"
	"fmt"
	"log"
	"time"
)

type App struct {
	Config      *config.Config
	PgClient    *postgres.Client
	Repo        *repository.Repository
	Server      *server.Server
	Storage     storage.FileStorage
	TaskManager tasks.TaskManager
}

func New() (*App, error) {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	// Подключаемся к PostgreSQL
	pgClient, err := postgres.New(&cfg.Postgres)
	if err != nil {
		return nil, err
	}

	// Создаем MinIO хранилище
	fileStorage, err := storage.NewMinIOStorage(
		cfg.MinIO.Endpoint,
		cfg.MinIO.AccessKey,
		cfg.MinIO.SecretKey,
		cfg.MinIO.BucketName,
		cfg.MinIO.UseSSL,
	)
	if err != nil {
		return nil, err
	}

	// Создаем репозиторий
	repo := repository.New(pgClient)

	// Создаем TaskManager
	taskManager := tasks.NewTaskManager(1) // Один воркер для последовательного выполнения

	// Создаем HTTP сервер
	srv := server.New(cfg, pgClient, repo, fileStorage, taskManager)

	return &App{
		Config:      cfg,
		PgClient:    pgClient,
		Repo:        repo,
		Server:      srv,
		Storage:     fileStorage,
		TaskManager: taskManager,
	}, nil
}

func (a *App) Start() error {
	// Запускаем TaskManager
	ctx := context.Background()
	if err := a.TaskManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start task manager: %w", err)
	}

	return a.Server.Start()
}

func (a *App) Close() error {
	// Останавливаем TaskManager
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.TaskManager.Stop(ctx); err != nil {
		log.Printf("Failed to stop task manager: %v", err)
	}

	return a.PgClient.Close()
}
