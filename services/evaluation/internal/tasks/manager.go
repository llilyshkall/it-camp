package tasks

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// taskItem элемент очереди задач с приоритетом
type taskItem struct {
	task      Task
	priority  int
	timestamp time.Time
	id        string
}

// taskManager реализация TaskManager
type taskManager struct {
	taskQueue   chan taskItem
	results     chan TaskResult
	stopChan    chan struct{}
	wg          sync.WaitGroup
	mu          sync.RWMutex
	stats       TaskStats
	isRunning   bool
	workerCount int
}

// NewTaskManager создает новый менеджер задач
func NewTaskManager(workerCount int) TaskManager {
	if workerCount <= 0 {
		workerCount = 1
	}

	return &taskManager{
		taskQueue:   make(chan taskItem, 1000), // буфер на 1000 задач
		results:     make(chan TaskResult, 1000),
		stopChan:    make(chan struct{}),
		workerCount: workerCount,
		stats: TaskStats{
			IsRunning: false,
		},
	}
}

// SubmitTask добавляет задачу в очередь
func (tm *taskManager) SubmitTask(task Task) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if !tm.isRunning {
		return fmt.Errorf("task manager is not running")
	}

	taskItem := taskItem{
		task:      task,
		priority:  task.GetPriority(),
		timestamp: time.Now(),
		id:        uuid.New().String(),
	}

	select {
	case tm.taskQueue <- taskItem:
		tm.stats.TotalTasks++
		tm.stats.PendingTasks++
		log.Printf("Task submitted: %s for project %d, priority: %d",
			task.GetProjectID(), task.GetPriority())
		return nil
	default:
		return fmt.Errorf("task queue is full")
	}
}

// Start запускает обработчик задач
func (tm *taskManager) Start(ctx context.Context) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.isRunning {
		return fmt.Errorf("task manager is already running")
	}

	tm.isRunning = true
	tm.stats.IsRunning = true

	// Запускаем воркеры
	for i := 0; i < tm.workerCount; i++ {
		tm.wg.Add(1)
		go tm.worker(ctx, i)
	}

	// Запускаем обработчик результатов
	tm.wg.Add(1)
	go tm.resultHandler(ctx)

	log.Printf("Task manager started with %d workers", tm.workerCount)
	return nil
}

// Stop останавливает обработчик задач
func (tm *taskManager) Stop(ctx context.Context) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if !tm.isRunning {
		return nil
	}

	log.Println("Stopping task manager...")

	// Сигнализируем остановку
	close(tm.stopChan)

	// Ждем завершения всех воркеров
	done := make(chan struct{})
	go func() {
		tm.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Task manager stopped gracefully")
	case <-ctx.Done():
		log.Println("Task manager stopped due to context timeout")
	}

	tm.isRunning = false
	tm.stats.IsRunning = false
	return nil
}

// GetStats возвращает статистику выполнения задач
func (tm *taskManager) GetStats() TaskStats {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return tm.stats
}

// worker основной воркер для выполнения задач
func (tm *taskManager) worker(ctx context.Context, workerID int) {
	defer tm.wg.Done()

	log.Printf("Worker %d started", workerID)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %d stopped due to context cancellation", workerID)
			return
		case <-tm.stopChan:
			log.Printf("Worker %d stopped due to stop signal", workerID)
			return
		case taskItem := <-tm.taskQueue:
			tm.executeTask(ctx, taskItem, workerID)
		}
	}
}

// executeTask выполняет конкретную задачу
func (tm *taskManager) executeTask(ctx context.Context, taskItem taskItem, workerID int) {
	startTime := time.Now()

	log.Printf("Worker %d executing task %s for project %d",
		workerID, taskItem.task.GetProjectID())

	// Выполняем задачу
	err := taskItem.task.Execute(ctx)

	duration := time.Since(startTime).Milliseconds()

	// Обновляем статистику
	tm.mu.Lock()
	tm.stats.PendingTasks--
	if err != nil {
		tm.stats.FailedTasks++
		log.Printf("Worker %d failed task %s for project %d: %v",
			workerID, taskItem.task.GetProjectID(), err)
	} else {
		tm.stats.CompletedTasks++
		log.Printf("Worker %d completed task %s for project %d in %dms",
			workerID, taskItem.task.GetProjectID(), duration)
	}
	tm.mu.Unlock()

	// Отправляем результат
	result := TaskResult{
		TaskID:    taskItem.id,
		ProjectID: taskItem.task.GetProjectID(),
		Success:   err == nil,
		Error:     err,
		Duration:  duration,
	}

	select {
	case tm.results <- result:
	default:
		log.Printf("Warning: results channel is full, dropping result for task %s", taskItem.id)
	}
}

// resultHandler обрабатывает результаты выполнения задач
func (tm *taskManager) resultHandler(ctx context.Context) {
	defer tm.wg.Done()

	log.Println("Result handler started")

	for {
		select {
		case <-ctx.Done():
			log.Println("Result handler stopped due to context cancellation")
			return
		case <-tm.stopChan:
			log.Println("Result handler stopped due to stop signal")
			return
		case result := <-tm.results:
			// Здесь можно добавить дополнительную обработку результатов
			// например, сохранение в БД, отправка уведомлений и т.д.
			if !result.Success {
				log.Printf("Task %s failed for project %d: %v",
					result.TaskID, result.ProjectID, result.Error)
			}
		}
	}
}
