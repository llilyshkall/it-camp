package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	deliv "remarks/internal/delivery"
	"remarks/internal/httputils"
	rep "remarks/internal/repository"
	uc "remarks/internal/usecase"
	"syscall"
	"time"
)

func main() {
	// Получаем порт из переменной окружения или используем по умолчанию
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Создаем HTTP сервер с таймаутами
	srv := &http.Server{
		Addr:              ":" + port,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	urlDB := "postgres://" + os.Getenv("TEST_POSTGRES_USER") + ":" + os.Getenv("TEST_POSTGRES_PASSWORD") + "@" + os.Getenv("TEST_DATABASE_HOST") + ":" + os.Getenv("DB_PORT") + "/" + os.Getenv("TEST_POSTGRES_DB")
	db, err := sql.Open("pgx", urlDB)
	if err != nil {
		log.Println("could not connect to database")
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Println("unable to reach database ", err)
	}
	log.Println("database is reachable")

	store := rep.NewSQLStore(db)

	usecase := uc.NewUsecase(store)

	handler := deliv.NewWebHandler(usecase)

	// Регистрируем обработчики
	http.HandleFunc("/health", httputils.HealthHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "remarks service", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
	})
	http.HandleFunc("/load_excel_registry", handler.LoadExcelRegistry)

	// Запускаем сервер в горутине
	go func() {
		log.Printf("Starting remarks service on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Настройка graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
