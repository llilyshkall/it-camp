.PHONY: help build test clean docker-build docker-compose-up docker-compose-down logs status health-check

# Загружаем переменные из .env файла
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Переменные для базы данных (с значениями по умолчанию)
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= evaluation_user
DB_PASSWORD ?= evaluation123
DB_NAME ?= evaluation_db

# Переменные для миграций (подключение к localhost)
MIGRATE_HOST ?= localhost
MIGRATE_PORT ?= 5432

# Переменные
SERVICE_NAME := evaluation
IMAGE_NAME := ghcr.io/llilyshkall/evaluation
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo 'dev')

help: ## Показать справку
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Собрать Go приложение
	@echo "Building $(SERVICE_NAME)..."
	cd services/$(SERVICE_NAME) && \
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-trimpath \
		-ldflags="-s -w -X main.version=$(VERSION)" \
		-o bin/$(SERVICE_NAME) \
		./cmd/$(SERVICE_NAME)

test: ## Запустить тесты
	@echo "Running tests..."
	cd services/$(SERVICE_NAME) && go test -v ./...

clean: ## Очистить артефакты сборки
	@echo "Cleaning build artifacts..."
	rm -rf services/$(SERVICE_NAME)/bin/

docker-build: ## Собрать Docker образ
	@echo "Building Docker image $(IMAGE_NAME):$(VERSION)..."
	docker build -t $(IMAGE_NAME):$(VERSION) -t $(IMAGE_NAME):latest services/$(SERVICE_NAME)

docker-compose-up: ## Запустить сервисы через docker-compose
	@echo "Starting services with docker-compose..."
	docker-compose up -d
	@echo "Waiting for PostgreSQL to be ready..."
	@until docker-compose exec -T postgres pg_isready -U evaluation_user -d evaluation_db; do \
		echo "Waiting for PostgreSQL..."; \
		sleep 2; \
	done
	@echo "PostgreSQL is ready!"
	@echo "Waiting for MinIO to be ready..."
	sleep 10
	@echo "Applying database migrations..."
	@if command -v migrate >/dev/null 2>&1; then \
		export PATH=$$PATH:$$(go env GOPATH)/bin; \
		make migrate-up; \
	else \
		echo "golang-migrate not found. Installing..."; \
		make migrate-install; \
		export PATH=$$PATH:$$(go env GOPATH)/bin; \
		make migrate-up; \
	fi
	@echo "Services started! PostgreSQL, MinIO, evaluation service, front and nginx are running."
	@echo "Frontend available at http://localhost:5000"
	@echo "API available at http://localhost:80/api"
	@echo "Service available at http://localhost:8081"
	@echo "MinIO Console available at http://localhost:9001 (minioadmin/minioadmin)"
	@echo "Use 'make logs' to see logs"
	@echo "Use 'make docker-compose-down' to stop"

docker-compose-down: ## Остановить docker-compose сервисы
	@echo "Stopping docker-compose services..."
	docker-compose down
	@echo "Services stopped!"

logs: ## Показать логи всех сервисов
	@echo "Showing logs for all services..."
	docker-compose logs -f

status: ## Показать статус всех контейнеров
	@echo "Checking Docker containers status..."
	docker-compose ps

health-check: ## Проверить здоровье сервиса
	@echo "Checking service health..."
	curl -f http://localhost:8081/health || echo "Service not accessible"

install-deps: ## Установить зависимости Go
	@echo "Installing Go dependencies..."
	cd services/$(SERVICE_NAME) && go mod download
	cd services/$(SERVICE_NAME) && go mod tidy

fmt: ## Форматировать код
	@echo "Formatting Go code..."
	cd services/$(SERVICE_NAME) && go fmt ./...

vet: ## Проверить код
	@echo "Running go vet..."
	cd services/$(SERVICE_NAME) && go vet ./...

# Команды для локальной разработки
dev: ## Запустить в режиме разработки
	@echo "Starting development mode..."
	cd services/$(SERVICE_NAME) && go run ./cmd/$(SERVICE_NAME)

# Установка golang-migrate
migrate-install: ## Установить golang-migrate
	@echo "Installing golang-migrate..."
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "golang-migrate installed successfully!"

# Команды для миграций
migrate-up: ## Применить все миграции
	@echo "Applying database migrations..."
	@if command -v migrate >/dev/null 2>&1; then \
		migrate -path services/evaluation/db/migration -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(MIGRATE_HOST):$(MIGRATE_PORT)/$(DB_NAME)?sslmode=disable" up; \
	else \
		echo "golang-migrate not found. Installing..."; \
		make migrate-install; \
		migrate -path services/evaluation/db/migration -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(MIGRATE_HOST):$(MIGRATE_PORT)/$(DB_NAME)?sslmode=disable" up; \
	fi

migrate-down: ## Откатить все миграции
	@echo "Rolling back all migrations..."
	@if command -v migrate >/dev/null 2>&1; then \
		echo "y" | migrate -path services/evaluation/db/migration -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(MIGRATE_HOST):$(MIGRATE_PORT)/$(DB_NAME)?sslmode=disable" down; \
	else \
		echo "golang-migrate not found. Installing..."; \
		make migrate-install; \
		echo "y" | migrate -path services/evaluation/db/migration -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(MIGRATE_HOST):$(MIGRATE_PORT)/$(DB_NAME)?sslmode=disable" down; \
	fi

migrate-force: ## Принудительно установить версию миграции
	@echo "Force setting migration version..."
	@if command -v migrate >/dev/null 2>&1; then \
		migrate -path services/evaluation/db/migration -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(MIGRATE_HOST):$(MIGRATE_PORT)/$(DB_NAME)?sslmode=disable" force $(version); \
	else \
		echo "golang-migrate not found. Installing..."; \
		make migrate-install; \
		migrate -path services/evaluation/db/migration -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(MIGRATE_HOST):$(MIGRATE_PORT)/$(DB_NAME)?sslmode=disable" force $(version); \
	fi

migrate-status: ## Показать статус миграций
	@echo "Checking migration status..."
	@if command -v migrate >/dev/null 2>&1; then \
		migrate -path services/evaluation/db/migration -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(MIGRATE_HOST):$(MIGRATE_PORT)/$(DB_NAME)?sslmode=disable" version; \
	else \
		echo "golang-migrate not found. Installing..."; \
		make migrate-install; \
		migrate -path services/evaluation/db/migration -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(MIGRATE_HOST):$(MIGRATE_PORT)/$(DB_NAME)?sslmode=disable" version; \
	fi

migrate-reset: ## Сбросить и пересоздать базу данных
	@echo "Resetting database..."
	docker-compose down -v
	docker-compose up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	sleep 15
	make migrate-up
	@echo "Database reset completed!"

# Создание новой миграции
migrate-create: ## Создать новую миграцию (usage: make migrate-create name=migration_name)
	@echo "Creating new migration: $(name)"
	@if command -v migrate >/dev/null 2>&1; then \
		migrate create -ext sql -dir services/evaluation/db/migration -seq $(name); \
	else \
		echo "golang-migrate not found. Installing..."; \
		make migrate-install; \
		migrate create -ext sql -dir services/evaluation/db/migration -seq $(name); \
	fi

# Команды для очистки
clean-all: ## Полная очистка
	@echo "Full cleanup..."
	make clean
	docker-compose down -v
	docker system prune -f
	@echo "Cleanup completed"

minio-status: ## Проверить статус MinIO
	@echo "Checking MinIO status..."
	@curl -f http://localhost:9000/minio/health/live || echo "MinIO not accessible"

# Команды для front сервиса
front-build: ## Собрать Docker образ для front
	@echo "Building front Docker image..."
	docker build -t evaluation-front:latest services/front

front-run: ## Запустить front локально
	@echo "Starting front service locally..."
	cd services/front && python run.py

front-install-deps: ## Установить зависимости для front
	@echo "Installing front dependencies..."
	cd services/front && pip install -r requirements.txt

front-clean: ## Очистить front артефакты
	@echo "Cleaning front artifacts..."
	cd services/front && make clean

