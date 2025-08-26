#!/bin/bash

# Скрипт для тестирования всех API ручек локально
# Убедитесь, что сервер запущен на порту 8081

BASE_URL="http://localhost:8081"
API_BASE="$BASE_URL/api"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для логирования
log() {
    echo -e "${BLUE}[$(date +'%H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}✓ $1${NC}"
}

error() {
    echo -e "${RED}✗ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Функция для проверки статуса ответа
check_status() {
    local status=$1
    local expected=$2
    local description=$3
    
    if [ "$status" -eq "$expected" ]; then
        success "$description - Status: $status"
        return 0
    else
        error "$description - Expected: $expected, Got: $status"
        return 1
    fi
}

# Функция для выполнения HTTP запроса
make_request() {
    local method=$1
    local url=$2
    local data=$3
    local description=$4
    
    log "Testing: $description"
    log "Request: $method $url"
    
    if [ -n "$data" ]; then
        log "Data: $data"
    fi
    
    # Выполняем запрос и сохраняем результат
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" -H "Content-Type: application/json" -d "$data")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url")
    fi
    
    # Разделяем ответ и статус код
    http_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    log "Response Status: $http_code"
    log "Response Body: $response_body"
    echo
    
    return $http_code
}

# Проверяем, что сервер запущен
log "Проверяем доступность сервера..."
if ! curl -s "$BASE_URL/health" > /dev/null; then
    error "Сервер не доступен на $BASE_URL"
    error "Убедитесь, что сервер запущен: go run ./cmd/evaluation"
    exit 1
fi
success "Сервер доступен на $BASE_URL"

echo
log "Начинаем тестирование API ручек..."
echo

# ========== HEALTH CHECK ==========
log "=== HEALTH CHECK ==="
make_request "GET" "$BASE_URL/health" "" "Health Check"
check_status $? 200 "Health Check"

# ========== PROJECTS ==========
log "=== PROJECTS ==="

# Создание проекта
log "--- Create Project ---"
create_project_data='{"name": "Test Project for API Testing"}'
make_request "POST" "$API_BASE/projects" "$create_project_data" "Create Project"
create_status=$?
check_status $create_status 201 "Create Project"

# Извлекаем ID созданного проекта из ответа
if [ $create_status -eq 201 ]; then
    project_id=$(echo "$response_body" | grep -o '"id":[0-9]*' | cut -d':' -f2)
    log "Created project ID: $project_id"
else
    warning "Не удалось создать проект, используем ID 1 для тестов"
    project_id=1
fi

# Получение списка проектов
log "--- List Projects ---"
make_request "GET" "$API_BASE/projects" "" "List Projects"
check_status $? 200 "List Projects"

# Получение конкретного проекта
log "--- Get Project ---"
make_request "GET" "$API_BASE/projects/$project_id" "" "Get Project"
check_status $? 200 "Get Project"

# ========== PROJECT FILES ==========
log "=== PROJECT FILES ==="

# Создаем временный тестовый файл
log "--- Create Test File ---"
test_file_content="This is a test file for API testing. Created at $(date)."
echo "$test_file_content" > test_file.txt
success "Created test file: test_file.txt"

# Загрузка файла в проект (documentation)
log "--- Upload Project File (documentation) ---"
upload_response=$(curl -s -w "\n%{http_code}" \
    -X POST \
    -F "file=@test_file.txt" \
    -F "type=documentation" \
    "$API_BASE/projects/$project_id/files")

upload_http_code=$(echo "$upload_response" | tail -n1)
upload_body=$(echo "$upload_response" | head -n -1)

log "Upload Response Status: $upload_http_code"
log "Upload Response Body: $upload_body"
check_status $upload_http_code 202 "Upload Project File (documentation)"

# Загрузка файла в проект (remarks)
log "--- Upload Project File (remarks) ---"
upload_response=$(curl -s -w "\n%{http_code}" \
    -X POST \
    -F "file=@test_file.txt" \
    -F "type=remarks" \
    "$API_BASE/projects/$project_id/files")

upload_http_code=$(echo "$upload_response" | tail -n1)
upload_body=$(echo "$upload_response" | head -n -1)

log "Upload Response Status: $upload_http_code"
log "Upload Response Body: $upload_body"
check_status $upload_http_code 202 "Upload Project File (remarks)"

# ========== ERROR CASES ==========
log "=== ERROR CASES ==="

# Неверный метод для health
log "--- Invalid Method for Health ---"
make_request "POST" "$BASE_URL/health" "" "Invalid Method for Health"
check_status $? 405 "Invalid Method for Health"

# Неверный метод для projects
log "--- Invalid Method for Projects ---"
make_request "PUT" "$API_BASE/projects" "" "Invalid Method for Projects"
check_status $? 405 "Invalid Method for Projects"

# Неверный метод для project files
log "--- Invalid Method for Project Files ---"
make_request "GET" "$API_BASE/projects/$project_id/files" "" "Invalid Method for Project Files"
check_status $? 405 "Invalid Method for Project Files"

# Несуществующий проект
log "--- Non-existent Project ---"
make_request "GET" "$API_BASE/projects/99999" "" "Non-existent Project"
check_status $? 404 "Non-existent Project"

# Неверный ID проекта
log "--- Invalid Project ID ---"
make_request "GET" "$API_BASE/projects/invalid" "" "Invalid Project ID"
check_status $? 400 "Invalid Project ID"

# Неверный путь для project files
log "--- Invalid Path for Project Files ---"
make_request "POST" "$API_BASE/projects/$project_id/invalid" "" "Invalid Path for Project Files"
check_status $? 400 "Invalid Path for Project Files"

# ========== SWAGGER DOCS ==========
log "=== SWAGGER DOCS ==="
log "--- Swagger UI ---"
make_request "GET" "$BASE_URL/docs/" "" "Swagger UI"
check_status $? 200 "Swagger UI"

# ========== CLEANUP ==========
log "=== CLEANUP ==="
rm -f test_file.txt
success "Removed test file"

echo
log "=== РЕЗУЛЬТАТЫ ТЕСТИРОВАНИЯ ==="
log "Все тесты завершены!"
log "Проверьте логи выше для деталей по каждому тесту"
echo
log "Для просмотра Swagger документации откройте: $BASE_URL/docs/"
log "Для просмотра API в браузере: $API_BASE/projects"
