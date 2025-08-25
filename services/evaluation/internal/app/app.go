package app

import (
	"evaluation/internal/config"
	"evaluation/internal/postgres"
	"evaluation/internal/repository"
	"evaluation/internal/server"
	"evaluation/internal/storage"
)

type App struct {
	Config   *config.Config
	PgClient *postgres.Client
	Repo     *repository.Repository
	Server   *server.Server
	Storage  storage.FileStorage
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

	// Создаем HTTP сервер
	srv := server.New(cfg, pgClient, repo, fileStorage)

	return &App{
		Config:   cfg,
		PgClient: pgClient,
		Repo:     repo,
		Server:   srv,
		Storage:  fileStorage,
	}, nil
}

func (a *App) Start() error {
	return a.Server.Start()
}

func (a *App) Close() error {
	return a.PgClient.Close()
}
