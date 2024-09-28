package main

import (
	"backup-to-minio/internal/backup"
	"backup-to-minio/internal/config"
	"log"
	"os"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
)

func main() {
	// Загрузка переменных окружения из .env файла
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file loaded, proceeding without environment variables.")
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

	// Создание нового планировщика
	s := gocron.NewScheduler(time.UTC)

	hasSchedule := false

	for _, backupItem := range cfg.Backups {
		if backupItem.Schedule != "" {
			// Запланировать выполнение резервного копирования
			_, err := s.Cron(backupItem.Schedule).Do(func(item config.ConfigBackup) func() {
				return func() {
					err := backup.ProcessBackup(item, bucketName)
					if err != nil {
						log.Printf("Failed to process backup for %s: %v", item.Name, err)
					} else {
						log.Printf("Backup for %s completed successfully.", item.Name)
					}
				}
			}(backupItem))
			if err != nil {
				log.Printf("Failed to schedule backup for %s: %v", backupItem.Name, err)
			} else {
				hasSchedule = true // Установить флаг, если расписание добавлено
			}
		} else {
			// Если расписание не указано, выполнить резервное копирование сразу
			err := backup.ProcessBackup(backupItem, bucketName)
			if err != nil {
				log.Printf("Failed to process backup for %s: %v", backupItem.Name, err)
			} else {
				log.Printf("Backup for %s completed successfully.", backupItem.Name)
			}
		}
	}

	// Запуск планировщика, если есть хотя бы одно расписание
	if hasSchedule {
		s.StartAsync()
		// Чтобы программа не завершалась, можно добавить бесконечный цикл
		select {}
	} else {
		log.Println("No scheduled backups found. Exiting program.")
		os.Exit(0) // Завершение программы
	}
}
