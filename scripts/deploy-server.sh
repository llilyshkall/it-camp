#!/bin/bash

# Скрипт для деплоя на сервер без Kubernetes
# Использование: ./scripts/deploy-server.sh [build|deploy|restart|status|logs]

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Переменные
SERVICE_NAME="remarks"
CONTAINER_NAME="remarks"
IMAGE_NAME="ghcr.io/yourusername/remarks"
PORT=80
CONTAINER_PORT=8080

# Функции для логирования
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка Docker
check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker не установлен"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        log_error "Docker демон не запущен"
        exit 1
    fi
}

# Сборка образа
build_image() {
    log_info "Сборка Docker образа..."
    
    if [ ! -f "services/$SERVICE_NAME/Dockerfile" ]; then
        log_error "Dockerfile не найден в services/$SERVICE_NAME/"
        exit 1
    fi
    
    docker build -t $IMAGE_NAME:latest services/$SERVICE_NAME/
    log_success "Образ собран успешно"
}

# Деплой на сервер
deploy() {
    log_info "Деплой на сервер..."
    
    # Остановка старого контейнера
    if docker ps -q -f name=$CONTAINER_NAME | grep -q .; then
        log_info "Остановка старого контейнера..."
        docker stop $CONTAINER_NAME
        docker rm $CONTAINER_NAME
    fi
    
    # Создание директории для данных
    mkdir -p /opt/$SERVICE_NAME
    
    # Запуск нового контейнера
    log_info "Запуск нового контейнера..."
    docker run -d \
        --name $CONTAINER_NAME \
        --restart unless-stopped \
        -p $PORT:$CONTAINER_PORT \
        -e ENVIRONMENT=production \
        -e PORT=$CONTAINER_PORT \
        -e TZ=UTC \
        -v /opt/$SERVICE_NAME:/app/data \
        --health-cmd="curl -f http://localhost:$CONTAINER_PORT/health || exit 1" \
        --health-interval=30s \
        --health-timeout=10s \
        --health-retries=3 \
        $IMAGE_NAME:latest
    
    # Ожидание готовности
    log_info "Ожидание готовности сервиса..."
    sleep 10
    
    # Проверка статуса
    if docker ps | grep -q $CONTAINER_NAME; then
        log_success "Контейнер запущен успешно"
    else
        log_error "Ошибка запуска контейнера"
        exit 1
    fi
    
    # Health check
    if curl -f "http://localhost:$PORT/health" &> /dev/null; then
        log_success "Health check прошел успешно"
    else
        log_warning "Health check не прошел, сервис может еще инициализироваться"
    fi
}

# Перезапуск сервиса
restart() {
    log_info "Перезапуск сервиса..."
    
    if docker ps -q -f name=$CONTAINER_NAME | grep -q .; then
        docker restart $CONTAINER_NAME
        log_success "Сервис перезапущен"
    else
        log_warning "Контейнер не найден, запускаю новый..."
        deploy
    fi
}

# Показать статус
status() {
    log_info "Статус сервиса:"
    echo
    
    # Статус контейнера
    if docker ps -q -f name=$CONTAINER_NAME | grep -q .; then
        docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep $CONTAINER_NAME
    else
        log_warning "Контейнер не запущен"
    fi
    
    echo
    
    # Health check
    if curl -f "http://localhost:$PORT/health" &> /dev/null; then
        log_success "Сервис доступен по адресу: http://localhost:$PORT"
        log_info "Health endpoint: http://localhost:$PORT/health"
    else
        log_warning "Сервис недоступен"
    fi
    
    echo
    
    # Использование ресурсов
    if docker ps -q -f name=$CONTAINER_NAME | grep -q .; then
        log_info "Использование ресурсов:"
        docker stats $CONTAINER_NAME --no-stream --format "table {{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}"
    fi
}

# Показать логи
logs() {
    if docker ps -q -f name=$CONTAINER_NAME | grep -q .; then
        log_info "Логи сервиса (Ctrl+C для выхода):"
        docker logs -f $CONTAINER_NAME
    else
        log_error "Контейнер не запущен"
        exit 1
    fi
}

# Остановка сервиса
stop() {
    log_info "Остановка сервиса..."
    
    if docker ps -q -f name=$CONTAINER_NAME | grep -q .; then
        docker stop $CONTAINER_NAME
        docker rm $CONTAINER_NAME
        log_success "Сервис остановлен"
    else
        log_warning "Контейнер не запущен"
    fi
}

# Очистка
cleanup() {
    log_info "Очистка..."
    
    # Остановка и удаление контейнера
    stop
    
    # Удаление образа
    if docker images | grep -q $IMAGE_NAME; then
        docker rmi $IMAGE_NAME:latest
        log_success "Образ удален"
    fi
    
    # Очистка неиспользуемых образов
    docker image prune -f
}

# Показать справку
help() {
    echo "Использование: $0 [команда]"
    echo ""
    echo "Команды:"
    echo "  build     - Собрать Docker образ"
    echo "  deploy    - Деплой на сервер"
    echo "  restart   - Перезапустить сервис"
    echo "  status    - Показать статус"
    echo "  logs      - Показать логи"
    echo "  stop      - Остановить сервис"
    echo "  cleanup   - Полная очистка"
    echo "  help      - Показать эту справку"
    echo ""
    echo "Примеры:"
    echo "  $0 build deploy    # Собрать и деплоить"
    echo "  $0 status          # Показать статус"
    echo "  $0 logs            # Показать логи"
}

# Основная логика
main() {
    if [ $# -eq 0 ]; then
        help
        exit 1
    fi
    
    check_docker
    
    for command in "$@"; do
        case $command in
            "build")
                build_image
                ;;
            "deploy")
                deploy
                ;;
            "restart")
                restart
                ;;
            "status")
                status
                ;;
            "logs")
                logs
                ;;
            "stop")
                stop
                ;;
            "cleanup")
                cleanup
                ;;
            "help")
                help
                ;;
            *)
                log_error "Неизвестная команда: $command"
                help
                exit 1
                ;;
        esac
    done
}

# Запуск основной функции
main "$@"
