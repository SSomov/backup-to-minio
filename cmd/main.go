package main

import (
	"log"
	"os"

	"backup-to-minio/internal/backup"
	"backup-to-minio/internal/config"

	"github.com/joho/godotenv"
)

func main() {
	// Загрузка переменных окружения из .env файла
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("No .env file loaded: %v", err)
	}

	// Чтение конфигурации
	cfg, err := config.LoadConfig("config.yml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Получение имени бакета из переменных окружения
	bucketName := os.Getenv("MINIO_BUCKET_NAME")
	if bucketName == "" {
		log.Fatalf("MINIO_BUCKET_NAME is not set")
	}

	// Процесс резервного копирования
	for _, backupItem := range cfg.Backups {
		err := backup.ProcessBackup(backupItem, bucketName)
		if err != nil {
			log.Printf("Failed to process backup for %s: %v", backupItem.Name, err)
		}
	}
}
