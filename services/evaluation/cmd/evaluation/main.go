package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"evaluation/internal/app"
)

func main() {
	// Инициализация приложения
	app, err := app.New()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	defer app.Close()

	// Запускаем сервер в горутине
	go func() {
		if err := app.Start(); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Настройка graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), app.Config.Server.ShutdownTimeout)
	defer cancel()

	if err := app.Server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
