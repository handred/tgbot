# Builder stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Установка зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd

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

# Рабочая директория
WORKDIR /app

# Копируем бинарник и устанавливаем владельца
COPY --from=builder --chown=appuser:appuser /app/main .

# Меняем пользователя
USER appuser

# Опционально: можно добавить entrypoint для автообновления yt-dlp
COPY docker/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

# Запуск приложения
CMD ["/app/main"]