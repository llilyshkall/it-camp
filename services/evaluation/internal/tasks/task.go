package tasks

import (
	"context"
)

// Task интерфейс для фоновых задач
type Task interface {
	// Execute выполняет задачу
	Execute(ctx context.Context) error

	// GetProjectID возвращает ID проекта, с которым работает задача
	GetProjectID() int32

	// GetPriority возвращает приоритет задачи (меньше = выше приоритет)
	GetPriority() int
}

// TaskResult результат выполнения задачи
type TaskResult struct {
	TaskID    string
	ProjectID int32
	Success   bool
	Error     error
	Duration  int64 // в миллисекундах
}

// TaskManager управляет выполнением фоновых задач
type TaskManager interface {
	// SubmitTask добавляет задачу в очередь
	SubmitTask(task Task) error

	// Start запускает обработчик задач
	Start(ctx context.Context) error

	// Stop останавливает обработчик задач
	Stop(ctx context.Context) error

	// GetStats возвращает статистику выполнения задач
	GetStats() TaskStats
}

// TaskStats статистика выполнения задач
type TaskStats struct {
	TotalTasks     int64
	CompletedTasks int64
	FailedTasks    int64
	PendingTasks   int
	IsRunning      bool
}
