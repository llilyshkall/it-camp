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

## 📁 Файловое хранилище (MinIO)

Сервис использует MinIO как S3-совместимое хранилище файлов для локальной разработки.

### Доступ к MinIO
- **API Endpoint**: http://localhost:9000
- **Web Console**: http://localhost:9001
- **Логин**: minioadmin / minioadmin

### Управление MinIO
```bash
# Открыть веб-консоль MinIO
make minio-console

# Проверить статус MinIO
make minio-status
```

### Bucket
- **Имя**: evaluation-files
- **Политика**: Публичный доступ на чтение
- **Структура**: Файлы сохраняются в папках по дате (YYYY-MM-DD/)

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
│   ├── repository/         # Слой доступа к данным
│   └── storage/            # Файловое хранилище (MinIO)
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

### MinIO конфигурация
- `MINIO_ENDPOINT` - адрес MinIO сервера (по умолчанию localhost:9000)
- `MINIO_ACCESS_KEY` - ключ доступа (по умолчанию minioadmin)
- `MINIO_SECRET_KEY` - секретный ключ (по умолчанию minioadmin)
- `MINIO_BUCKET_NAME` - имя bucket (по умолчанию evaluation-files)
- `MINIO_USE_SSL` - использование SSL (по умолчанию false)

## 📊 Мониторинг

- Health check: `http://localhost:8081/health`
- MinIO health: `http://localhost:9000/minio/health/live`

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
