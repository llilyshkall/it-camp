# IT-Camp Evaluation Service

Сервис для оценки проектов с поддержкой замечаний и файлов.

## 🚀 Быстрый старт

### Запуск сервисов
```bash
# Запустить все сервисы с автоматическим применением миграций
make docker-compose-up

# Проверить статус
make status

# Проверить здоровье сервиса
make health-check
```

### Остановка сервисов
```bash
make docker-compose-down
```

## 🗄️ Управление базой данных

### Миграции
```bash
# Применить все миграции
make migrate-up

# Откатить все миграции
make migrate-down

# Проверить статус миграций
make migrate-status

# Создать новую миграцию
make migrate-create name=add_new_field

# Сбросить базу данных
make migrate-reset
```

### Подробная документация по миграциям
См. [MIGRATION_GUIDE.md](MIGRATION_GUIDE.md)

## 📚 API документация

Примеры запросов к API см. в [API_EXAMPLES.md](API_EXAMPLES.md)

## 🛠️ Разработка

### Сборка
```bash
make build
```

### Тестирование
```bash
make test
```

### Локальная разработка
```bash
make dev
```

### Логи
```bash
make logs
```

## 📁 Структура проекта

```
services/evaluation/
├── cmd/evaluation/          # Точка входа
├── internal/                # Внутренняя логика
│   ├── app/                # Инициализация приложения
│   ├── config/             # Конфигурация
│   ├── handler/            # HTTP обработчики
│   ├── postgres/           # Работа с базой данных
│   └── repository/         # Слой доступа к данным
├── db/                     # База данных
│   ├── migration/          # Миграции
│   └── query/              # SQL запросы
└── Dockerfile              # Docker образ
```

## 🔧 Конфигурация

Сервис использует переменные окружения из файла `.env`:

- `DB_PASSWORD` - пароль для базы данных
- `PORT` - порт сервиса (по умолчанию 8080)
- `ENVIRONMENT` - окружение (development/production)

## 📊 Мониторинг

- Health check: `http://localhost:8081/health`
- API endpoints: `http://localhost:8081/api/*`
- PostgreSQL: `localhost:5432`

## 🚨 Troubleshooting

### Проблемы с миграциями
```bash
# Проверить статус
make migrate-status

# Принудительно исправить версию
make migrate-force version=1

# Полный сброс
make migrate-reset
```

### Проблемы с Docker
```bash
# Проверить статус контейнеров
make status

# Посмотреть логи
make logs

# Перезапустить сервисы
make docker-compose-down && make docker-compose-up
```
