# Build stage
FROM golang:1.24.5-alpine AS builder

# Устанавливаем необходимые пакеты
RUN apk add --no-cache ca-certificates tzdata

# Создаем пользователя для сборки
RUN adduser -D -g '' appuser

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" \
    -o /app/evaluation \
    ./cmd/evaluation

# Final stage
FROM gcr.io/distroless/static:nonroot

# Копируем сертификаты и временную зону
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Копируем бинарник
COPY --from=builder /app/evaluation /evaluation

# Устанавливаем переменные окружения
ENV TZ=UTC
ENV PORT=8080

# Открываем порт
EXPOSE 8080

# Запускаем приложение
ENTRYPOINT ["/evaluation"]
