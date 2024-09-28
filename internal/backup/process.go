package backup

import (
	// Импортируем пакет config
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
	tarName := GenerateTarName(backupItem.Name)
	tmpFilePath := filepath.Join("/tmp", tarName)

	var err error

	// В зависимости от типа выполняем резервное копирование папки или Docker volume
	if backupItem.Type == "folder" {
		_, err = TarFolder(backupItem.Source, tmpFilePath)
	} else {
		return fmt.Errorf("unsupported backup type: %s", backupItem.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to create backup for %s: %w", backupItem.Source, err)
	}

	// Определение пути в бакете, если указан path-save
	objectName := tarName
	if backupItem.PathSave != "" {
		objectName = filepath.Join(backupItem.PathSave, tarName)
	}

	// Загрузка архива в MinIO
	err = minio.UploadToMinio(bucketName, objectName, tmpFilePath)
	if err != nil {
		return fmt.Errorf("failed to upload %s to MinIO: %w", objectName, err)
	}

	// Удаление временного файла
	err = os.Remove(tmpFilePath)
	if err != nil {
		log.Printf("Failed to remove temporary file %s: %v", tmpFilePath, err)
	} else {
		log.Printf("Temporary file %s removed", tmpFilePath)
	}

	log.Printf("File %s successfully uploaded to MinIO", objectName)
	return nil
}
