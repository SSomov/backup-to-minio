package main

import (
	"backup-to-minio/internal/backup"
	"backup-to-minio/internal/config"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
)

func main() {
	// Загрузка переменных окружения
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using only system environment variables")
	}

	// Загрузка конфигурации
	cfg, err := config.LoadConfig("config.yml")
	if err != nil {
		log.Fatalf("Config loading failed: %v", err)
	}

	// Получение имени бакета
	bucketName := os.Getenv("MINIO_BUCKET_NAME")
	if bucketName == "" {
		log.Fatal("MINIO_BUCKET_NAME environment variable is required")
	}

	// Инициализация планировщика
	scheduler := gocron.NewScheduler(time.UTC)
	hasScheduledJobs := false

	// Обработка CTRL+C для graceful shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// Обработка всех бэкапов
	for i := range cfg.Backups {
		item := cfg.Backups[i]

		if item.Schedule != "" {
			// Запланированное выполнение
			_, err := scheduler.Cron(item.Schedule).Do(func(c *config.BackupConfig, b config.ConfigBackup, bucket string) {
				log.Printf("Starting scheduled backup: %s", b.Name)
				if err := backup.ProcessBackup(c, b, bucket); err != nil {
					log.Printf("Backup %s failed: %v", b.Name, err)
				}
			}, cfg, item, bucketName)

			if err != nil {
				log.Printf("Failed to schedule %s: %v", item.Name, err)
				continue
			}
			hasScheduledJobs = true
		} else {
			// Немедленное выполнение
			log.Printf("Starting immediate backup: %s", item.Name)
			if err := backup.ProcessBackup(cfg, item, bucketName); err != nil {
				log.Printf("Backup %s failed: %v", item.Name, err)
			}
		}
	}

	// Запуск планировщика если есть задания
	if hasScheduledJobs {
		scheduler.StartAsync()
		log.Println("Scheduler started. Press CTRL+C to exit")

		// Ожидание сигнала завершения
		<-stopChan
		log.Println("Shutting down scheduler...")
		scheduler.Stop()
		log.Println("Scheduler stopped")
	} else {
		log.Println("No scheduled backups found. Exiting")
	}
}
