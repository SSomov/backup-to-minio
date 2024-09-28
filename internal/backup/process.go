package backup

import (
	"backup-to-minio/internal/config"
	"backup-to-minio/internal/minio"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// ProcessBackup обрабатывает резервное копирование для каждого элемента из конфигурации
func ProcessBackup(backupItem config.ConfigBackup, bucketName string) error {
	// Генерация имени архива с текущей датой и временем
	var filePath string
	var err error

	// В зависимости от типа выполняем резервное копирование папки, PostgreSQL или MySQL базы данных
	switch backupItem.Type {
	case "folder":
		// Для папки создаем tar-архив
		tarName := GenerateTarName(backupItem.Name)
		tmpFilePath := filepath.Join("/tmp", tarName)
		filePath, err = TarFolder(backupItem.Source, tmpFilePath)

	case "postgres":
		// Для PostgreSQL делаем дамп базы данных
		filePath, err = BackupPostgres(backupItem.Source, "/tmp")

	case "mysql":
		// Для MySQL делаем дамп базы данных
		filePath, err = BackupMySQL(backupItem.Source, "/tmp")

	default:
		return fmt.Errorf("unsupported backup type: %s", backupItem.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to create backup for %s: %w", backupItem.Source, err)
	}

	// Определение пути в бакете, если указан path-save
	objectName := filepath.Base(filePath)
	if backupItem.PathSave != "" {
		objectName = filepath.Join(backupItem.PathSave, objectName)
	}

	// Загрузка архива в MinIO
	err = minio.UploadToMinio(bucketName, objectName, filePath)
	if err != nil {
		return fmt.Errorf("failed to upload %s to MinIO: %w", objectName, err)
	}

	// Удаление временного файла
	err = os.Remove(filePath)
	if err != nil {
		log.Printf("Failed to remove temporary file %s: %v", filePath, err)
	} else {
		log.Printf("Temporary file %s removed", filePath)
	}

	log.Printf("File %s successfully uploaded to MinIO", objectName)
	return nil
}
