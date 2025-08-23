# IT-Camp Project

Микросервисное приложение с автоматическим развертыванием через GitHub Actions.

## 🚀 Быстрый старт

### Локальная разработка

```bash
# Клонировать репозиторий
git clone <your-repo-url>
cd it-camp

# Установить зависимости
make install-deps

# Запустить тесты
make test

# Собрать и запустить локально
make build
make docker-compose-up

# Проверить здоровье сервиса
make health-check
```

### Развертывание на сервер

#### Автоматический деплой через GitHub Actions
1. Настройте GitHub Secrets (см. SETUP-SIMPLE.md)
2. Сделайте push в main ветку
3. GitHub Actions автоматически задеплоит на сервер

#### Ручной деплой
```bash
# Собрать и деплоить
./scripts/deploy-server.sh build deploy

# Проверить статус
./scripts/deploy-server.sh status

# Посмотреть логи
./scripts/deploy-server.sh logs
```

## 📁 Структура проекта

```
it-camp/
├── .github/workflows/          # GitHub Actions CI/CD
├── scripts/                    # Скрипты для деплоя
│   └── deploy-server.sh       # Скрипт деплоя на сервер
├── services/
│   └── remarks/                # Go микросервис
│       ├── cmd/remarks/
│       ├── internal/httputils/
│       ├── Dockerfile
│       ├── go.mod
│       └── .dockerignore
├── docker-compose.yaml         # Локальная разработка
├── Makefile                    # Команды для разработки
├── README.md
└── SETUP-SIMPLE.md            # Инструкция по настройке
```

## 🔧 Настройка для продакшна

### 1. GitHub Container Registry

1. Создать Personal Access Token в GitHub:
   - Settings → Developer settings → Personal access tokens
   - Выбрать `write:packages` scope

2. Обновить `Makefile`:
   ```makefile
   IMAGE_NAME := ghcr.io/ВАШ_USERNAME/remarks
   ```

### 2. Подготовка сервера

```bash
# Установить Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
sudo systemctl enable docker
sudo systemctl start docker

# Создать SSH ключ для GitHub Actions
ssh-keygen -t rsa -b 4096 -C "github-actions"
cat ~/.ssh/id_rsa.pub >> ~/.ssh/authorized_keys
```

### 3. Настройка GitHub Secrets

В репозитории GitHub → Settings → Secrets and variables → Actions добавьте:
- `SERVER_HOST`: IP вашего сервера
- `SERVER_USER`: пользователь на сервере
- `SERVER_SSH_KEY`: приватный SSH ключ
- `SERVER_PORT`: SSH порт (обычно 22)

## 🚀 CI/CD Pipeline

### Автоматический деплой

1. При пуше в `main` ветку автоматически:
   - Запускаются тесты
   - Собирается Docker образ
   - Образ пушится в GitHub Container Registry
   - Происходит деплой на сервер

2. Ручной запуск:
   - GitHub → Actions → Deploy to Server (Docker) → Run workflow

### Мониторинг деплоя

```bash
# Проверить статус
make status

# Посмотреть логи
make logs

# Health check
make health-check
```

## 📊 Мониторинг

### Логи

```bash
# Логи приложения
make logs

# Статистика контейнеров
make monitor
```

### Health Check

```bash
# Проверка здоровья
make health-check

# Прямая проверка
curl http://localhost:8081/health
```

## 🔒 Безопасность

### Firewall

```bash
# Открыть только нужные порты
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw enable
```

### SSH

```bash
# Отключить парольную аутентификацию
sudo nano /etc/ssh/sshd_config
# PasswordAuthentication no
sudo systemctl restart sshd
```

## 🧪 Тестирование

### Локальные тесты

```bash
# Запустить тесты
make test

# Тесты с покрытием
make test-coverage

# Проверить код
make vet

# Форматирование
make fmt
```

### Интеграционные тесты

```bash
# Запустить сервис
make docker-compose-up

# Тест health endpoint
curl http://localhost:8081/health

# Тест основного endpoint
curl http://localhost:8081/
```

## 📝 Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `PORT` | Порт для HTTP сервера | `8080` |
| `ENVIRONMENT` | Окружение (dev/prod) | `development` |
| `TZ` | Временная зона | `UTC` |
| `LOG_LEVEL` | Уровень логирования | `info` |

## 🔄 Обновление

### Обновление кода

```bash
# Pull изменений
git pull origin main

# Пересобрать образ
make docker-build

# Обновить на сервере
./scripts/deploy-server.sh deploy
```

### Обновление зависимостей

```bash
# Обновить Go модули
cd services/remarks
go get -u ./...
go mod tidy

# Обновить Docker образы
docker pull golang:1.22.5-alpine
docker pull gcr.io/distroless/static:nonroot
```

## 🆘 Troubleshooting

### Проблемы с деплоем

```bash
# Проверить статус контейнера
make status

# Посмотреть логи
make logs

# Перезапустить
./scripts/deploy-server.sh restart
```

### Проблемы с образом

```bash
# Проверить локальные образы
docker images | grep remarks

# Пересобрать образ
make docker-build
```

### Проблемы с сетью

```bash
# Проверить порты
netstat -tlnp | grep 8081

# Проверить firewall
sudo ufw status
```

## 📚 Полезные команды

```bash
# Показать справку
make help

# Запуск в режиме разработки
make dev

# Запуск через Docker
make dev-docker

# Проверка здоровья
make health-check

# Очистка
make clean-all
```

## 🤝 Вклад в проект

1. Fork репозитория
2. Создать feature branch (`git checkout -b feature/amazing-feature`)
3. Commit изменения (`git commit -m 'Add amazing feature'`)
4. Push в branch (`git push origin feature/amazing-feature`)
5. Создать Pull Request

## 📄 Лицензия

Этот проект лицензирован под MIT License - см. файл [LICENSE](LICENSE) для деталей.

## 📞 Поддержка

Если у вас есть вопросы или проблемы:

1. Создать Issue в GitHub
2. Проверить документацию
3. Обратиться к команде разработки
