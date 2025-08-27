#!/bin/sh
# entrypoint.sh

# Автообновление yt-dlp
echo "Updating yt-dlp..."
yt-dlp --update || echo "Failed to update yt-dlp, continuing..."

# Запуск основного приложения
exec "$@"