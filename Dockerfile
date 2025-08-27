# Builder stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Установка зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода (включая docker/)
COPY . .

# Сборка приложения — УДАЛЁН -installsuffix cgo
RUN CGO_ENABLED=0 GOOS=linux go build -a -o main .

# Final stage
FROM alpine:latest

# Установка необходимых пакетов
RUN apk --no-cache add \
    ca-certificates \
    bash \
    ffmpeg \
    yt-dlp \
    wget

# Создаём непривилегированного пользователя
RUN adduser -D -s /bin/sh appuser

RUN mkdir -p /app/downloads/video /app/downloads/audio /app/downloads/image && \
    chown -R appuser:appuser /app/downloads

# Рабочая директория
WORKDIR /app

# Копируем бинарник
COPY --from=builder --chown=appuser:appuser /app/main .

# Копируем entrypoint.sh из builder
COPY --from=builder --chown=appuser:appuser /app/docker/entrypoint.sh /usr/local/bin/entrypoint.sh

# Делаем исполняемым
RUN chmod +x /usr/local/bin/entrypoint.sh

# Меняем пользователя
USER appuser

# Запуск
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["./main"]