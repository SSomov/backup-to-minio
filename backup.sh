#!/bin/bash

# Загрузка переменных из .env файла
set -o allexport
source /backup.env
set +o allexport

# Чтение файла конфигурации
CONFIG_FILE="/backup.yml"

if [ ! -f "$CONFIG_FILE" ]; then
    echo "Configuration file not found!"
    exit 1
fi

# Установка Minio alias с авторизацией
mc alias set myminio $MINIO_ENDPOINT $MINIO_ACCESS_KEY $MINIO_SECRET_KEY

# Функция обработки и бэкапа
backup() {
    local name=$1
    local source=$2
    local type=$3
    local bucket=$4
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local archive_name="${name}_backup_$timestamp.tar.gz"

    if [ "$type" == "folder" ]; then
        tar -czf /tmp/$archive_name -C "$source" .
    elif [ "$type" == "volume" ]; then
        docker run --rm -v ${source}:/data -v $(pwd):/backup alpine \
            tar -czf /backup/$archive_name -C /data .
    fi

    mc cp /tmp/$archive_name myminio/$bucket/
    rm /tmp/$archive_name
}

# Парсинг YAML и выполнение бэкапов
while IFS= read -r line; do
    eval "$line"
done < <(yq e '.backups[] | .name, .source, .type, .bucket' "$CONFIG_FILE")

backup "$name" "$source" "$type" "$bucket"
