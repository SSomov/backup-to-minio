package minio

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// UploadParams содержит параметры для загрузки в MinIO
type UploadParams struct {
	Project    string // Имя проекта (ybex)
	BucketName string // Название бакета
	ObjectPath string // Путь к объекту в бакете (без проекта)
	FilePath   string // Локальный путь к файлу
}

// UploadToMinio загружает файл в MinIO с учетом структуры проекта
func UploadToMinio(params UploadParams) error {
	// Валидация параметров
	if params.Project == "" || params.BucketName == "" || params.ObjectPath == "" || params.FilePath == "" {
		return fmt.Errorf("all upload parameters must be specified")
	}

	// Получение настроек из окружения
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	// Инициализация клиента
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return fmt.Errorf("minio initialization failed: %v", err)
	}

	ctx := context.Background()

	// Проверка и создание бакета при необходимости
	exists, err := minioClient.BucketExists(ctx, params.BucketName)
	if err != nil {
		return fmt.Errorf("bucket check failed: %v", err)
	}

	if !exists {
		err = minioClient.MakeBucket(ctx, params.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("bucket creation failed: %v", err)
		}
	}

	// Формирование полного пути в бакете
	fullObjectPath := filepath.Join(params.Project, params.ObjectPath)

	// Нормализация путей для MinIO
	fullObjectPath = strings.ReplaceAll(fullObjectPath, string(filepath.Separator), "/")

	// Загрузка файла
	_, err = minioClient.FPutObject(
		ctx,
		params.BucketName,
		fullObjectPath,
		params.FilePath,
		minio.PutObjectOptions{
			UserMetadata: map[string]string{
				"x-amz-acl": "private",
				"Project":   params.Project,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("file upload failed: %v", err)
	}

	fmt.Printf("Successfully uploaded %s to %s/%s\n",
		params.FilePath,
		params.BucketName,
		fullObjectPath,
	)

	return nil
}
