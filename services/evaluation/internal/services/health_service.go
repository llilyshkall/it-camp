package services

import (
	"context"
	"evaluation/internal/postgres"
	"time"
)

// healthService реализация HealthService
type healthService struct {
	pgClient *postgres.Client
}

// NewHealthService создает новый экземпляр HealthService
func NewHealthService(pgClient *postgres.Client) HealthService {
	return &healthService{
		pgClient: pgClient,
	}
}

// CheckHealth проверяет состояние сервиса
func (s *healthService) CheckHealth(ctx context.Context) (*HealthResponse, error) {
	// Проверяем состояние базы данных
	dbStatus := "healthy"
	if err := s.pgClient.HealthCheck(ctx); err != nil {
		dbStatus = "unhealthy"
	}

	response := &HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().Format(time.RFC3339),
		Database:  dbStatus,
	}

	return response, nil
}
