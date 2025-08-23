.PHONY: help build test clean docker-build docker-compose-up docker-compose-down logs status health-check

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
	@echo "Services started! PostgreSQL and evaluation service are running."
	@echo "Service available at http://localhost:8081"
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

# Команды для очистки
clean-all: ## Полная очистка
	@echo "Full cleanup..."
	make clean
	docker-compose down -v
	docker system prune -f
	@echo "Cleanup completed"
