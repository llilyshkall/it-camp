#!/bin/bash

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Тестирование GET /api/projects/{id} ===${NC}"
echo

# Проверяем, запущен ли сервис
if ! curl -s -f http://localhost:8081/health > /dev/null 2>&1; then
    echo -e "${RED}Ошибка: Сервис не запущен на http://localhost:8081${NC}"
    echo -e "${YELLOW}Запустите сервисы командой: make docker-compose-up${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Сервис доступен${NC}"
echo

# 1. Создаем тестовый проект
echo -e "${YELLOW}1. Создание тестового проекта:${NC}"
project_response=$(curl -s -X POST http://localhost:8081/api/projects \
    -H "Content-Type: application/json" \
    -d '{"name": "Тестовый проект для GET"}')
echo "$project_response" | jq . 2>/dev/null || echo "$project_response"
echo

# 2. Получаем проект по ID
echo -e "${YELLOW}2. Получение проекта по ID $project_id:${NC}"
get_project_response=$(curl -s http://localhost:8081/api/projects/1)
echo "$get_project_response" | jq . 2>/dev/null || echo "$get_project_response"
echo

# 3. Тест с несуществующим ID
echo -e "${YELLOW}3. Тест с несуществующим ID (999):${NC}"
not_found_response=$(curl -s http://localhost:8081/api/projects/999)
echo "$not_found_response" | jq . 2>/dev/null || echo "$not_found_response"
echo

# 4. Тест с некорректным ID (буквы)
echo -e "${YELLOW}4. Тест с некорректным ID (abc):${NC}"
invalid_id_response=$(curl -s http://localhost:8081/api/projects/abc)
echo "$invalid_id_response" | jq . 2>/dev/null || echo "$invalid_id_response"
echo

# 5. Тест с неправильным HTTP методом (POST)
echo -e "${YELLOW}5. Тест с неправильным HTTP методом (POST):${NC}"
wrong_method_response=$(curl -s -X POST http://localhost:8081/api/projects/$project_id \
    -H "Content-Type: application/json" \
    -d '{"name": "Тест"}')
echo "$wrong_method_response" | jq . 2>/dev/null || echo "$wrong_method_response"
echo

echo -e "${GREEN}=== Тестирование завершено ===${NC}"
echo -e "${BLUE}Swagger документация: http://localhost:8081/api/docs/${NC}"
