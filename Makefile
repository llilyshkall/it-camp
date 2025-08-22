.PHONY: help build test clean docker-build docker-run docker-push docker-compose-up docker-compose-down deploy-server

# Переменные
SERVICE_NAME := remarks
IMAGE_NAME := ghcr.io/llilyshkall/remarks
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

test-coverage: ## Запустить тесты с покрытием
	@echo "Running tests with coverage..."
	cd services/$(SERVICE_NAME) && go test -v -coverprofile=coverage.out ./...
	cd services/$(SERVICE_NAME) && go tool cover -html=coverage.out

clean: ## Очистить артефакты сборки
	@echo "Cleaning build artifacts..."
	rm -rf services/$(SERVICE_NAME)/bin/
	rm -rf services/$(SERVICE_NAME)/coverage.out

docker-build: ## Собрать Docker образ
	@echo "Building Docker image $(IMAGE_NAME):$(VERSION)..."
	docker build -t $(IMAGE_NAME):$(VERSION) -t $(IMAGE_NAME):latest services/$(SERVICE_NAME)

docker-run: ## Запустить Docker контейнер
	@echo "Running Docker container..."
	docker run -d --name $(SERVICE_NAME)-dev -p 8081:8080 \
		-e PORT=8080 \
		-e ENVIRONMENT=development \
		$(IMAGE_NAME):latest

docker-stop: ## Остановить Docker контейнер
	@echo "Stopping Docker container..."
	docker stop $(SERVICE_NAME)-dev || true
	docker rm $(SERVICE_NAME)-dev || true

docker-push: ## Отправить Docker образ в registry
	@echo "Pushing Docker image to registry..."
	docker push $(IMAGE_NAME):$(VERSION)
	docker push $(IMAGE_NAME):latest

docker-compose-up: ## Запустить через docker-compose
	@echo "Starting services with docker-compose..."
	docker-compose up -d

docker-compose-down: ## Остановить docker-compose
	@echo "Stopping docker-compose services..."
	docker-compose down

deploy-server: ## Деплой на сервер (требует настройки)
	@echo "Deploying to server..."
	@echo "Make sure you have configured GitHub Secrets first!"
	@echo "See SETUP-SIMPLE.md for instructions"
	@echo ""
	@echo "Or deploy manually:"
	@echo "  ./scripts/deploy-server.sh build deploy"

status: ## Показать статус Docker контейнеров
	@echo "Checking Docker containers status..."
	docker ps -a | grep $(SERVICE_NAME) || echo "No containers found"

logs: ## Показать логи Docker контейнера
	@echo "Showing logs for $(SERVICE_NAME)..."
	docker logs -f $(SERVICE_NAME)-dev || echo "Container not running"

port-forward: ## Проброс портов для локального доступа
	@echo "Setting up port forward..."
	@echo "Service should be available at http://localhost:8081"

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

lint: ## Запустить линтер (если установлен golangci-lint)
	@echo "Running linter..."
	cd services/$(SERVICE_NAME) && golangci-lint run

# Команды для локальной разработки
dev: ## Запустить в режиме разработки
	@echo "Starting development mode..."
	cd services/$(SERVICE_NAME) && go run ./cmd/$(SERVICE_NAME)

dev-docker: ## Запустить в Docker режиме разработки
	@echo "Starting Docker development mode..."
	make docker-compose-up
	@echo "Service available at http://localhost:8081"
	@echo "Use 'make logs' to see logs"
	@echo "Use 'make docker-compose-down' to stop"

# Команды для мониторинга
monitor: ## Показать статистику Docker контейнеров
	@echo "Docker containers statistics:"
	docker stats --no-stream $(SERVICE_NAME)-dev 2>/dev/null || echo "Container not running"

# Команды для очистки
clean-all: ## Полная очистка
	@echo "Full cleanup..."
	make clean
	docker system prune -f
	@echo "Cleanup completed"
