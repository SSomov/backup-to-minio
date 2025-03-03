package backup

import (
	"backup-to-minio/internal/config"
	"backup-to-minio/internal/minio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// ProcessBackup обрабатывает резервное копирование для каждого элемента из конфигурации
func ProcessBackup(cfg *config.BackupConfig, backupItem config.ConfigBackup, bucketName string) error {
	// Генерация временной метки для имен файлов
	timestamp := time.Now().Format("2006-01-02T15-04-05Z")

	// Создаем временную директорию для бэкапов
	tmpDir := filepath.Join(os.TempDir(), "backups", cfg.Project)
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	var filePath string
	var err error

	// Обработка разных типов бэкапов
	switch backupItem.Type {
	case "folder":
		tarName := fmt.Sprintf("%s-%s.tar.gz", backupItem.Name, timestamp)
		tmpFilePath := filepath.Join(tmpDir, tarName)
		filePath, err = TarFolder(backupItem.Source, tmpFilePath)
	
	case "mongodb":
		filePath, err = BackupMongoDB(backupItem.Source, tmpDir)
		if err == nil {
			log.Printf("MongoDB backup created: %s", filePath)
		}

	case "postgres":
		filePath, err = BackupPostgres(backupItem.Source, tmpDir)
		if err == nil {
			log.Printf("Postgres backup created: %s", filePath)
		}

	case "mysql":
		filePath, err = BackupMySQL(backupItem.Source, tmpDir)
		if err == nil {
			log.Printf("MySQL backup created: %s", filePath)
		}

	default:
		return fmt.Errorf("unsupported backup type: %s", backupItem.Type)
	}

	if err != nil {
		return fmt.Errorf("backup failed for %s: %w", backupItem.Name, err)
	}

	// Формирование пути в бакете
	objectPath := backupItem.PathSave
	if objectPath == "" {
		objectPath = backupItem.Name
	}
	objectName := filepath.Join(objectPath, filepath.Base(filePath))

	// Параметры для загрузки в MinIO
	uploadParams := minio.UploadParams{
		Project:    cfg.Project,
		BucketName: bucketName,
		ObjectPath: objectName,
		FilePath:   filePath,
	}

	// Загрузка в MinIO
	if err := minio.UploadToMinio(uploadParams); err != nil {
		return fmt.Errorf("minio upload failed: %w", err)
	}

	// Очистка временных файлов
	if err := os.Remove(filePath); err != nil {
		log.Printf("Warning: failed to remove temp file %s: %v", filePath, err)
	}
	log.Printf("Successfully processed backup: %s", backupItem.Name)

	return nil
}
