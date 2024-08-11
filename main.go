package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gopkg.in/yaml.v3"
)

// BackupConfig представляет конфигурацию резервного копирования
type BackupConfig struct {
	Backups []struct {
		Name     string `yaml:"name"`
		Source   string `yaml:"source"`
		Type     string `yaml:"type"`
		PathSave string `yaml:"path-save,omitempty"` // Необязательный параметр
	} `yaml:"backups"`
}

// tarFolder сжимает все файлы и папки в указанной директории и возвращает путь к созданному архиву
func tarFolder(source, target string) (string, error) {
	tarfile, err := os.Create(target)
	if err != nil {
		return "", err
	}
	defer tarfile.Close()

	gzw := gzip.NewWriter(tarfile)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	baseDir := filepath.Base(source)

	// Создаем корневую папку внутри архива
	if err := tw.WriteHeader(&tar.Header{
		Name:     baseDir + "/",
		Mode:     0755,
		Typeflag: tar.TypeDir,
	}); err != nil {
		return "", err
	}

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, relPath)
		if err != nil {
			return err
		}
		header.Name = filepath.Join(baseDir, relPath)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(tw, file)
		return err
	})

	if err != nil {
		return "", err
	}

	return tarfile.Name(), nil
}

// createDockerVolumeBackup создает резервную копию Docker volume в формате tar.gz с помощью контейнера BusyBox
func createDockerVolumeBackup(volumeName, target string) error {
	// Формируем имя файла архива
	archiveName := filepath.Base(target)

	// Команда для создания резервной копии
	cmd := exec.Command("docker", "run", "--rm",
		"-v", fmt.Sprintf("%s:/vackup-volume", volumeName),
		"-v", fmt.Sprintf("%s:/vackup", filepath.Dir(target)),
		"busybox", "tar", "-zcvf", fmt.Sprintf("/vackup/%s", archiveName), "/vackup-volume")

	// Запуск команды
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create Docker volume backup: %v, output: %s", err, string(output))
	}

	return nil
}

// generateTarName генерирует имя файла архива с текущей датой и временем в формате ISO 8601
func generateTarName(baseName string) string {
	now := time.Now().UTC()
	timestamp := now.Format("2006-01-02T15:04:05Z")
	return fmt.Sprintf("%s_%s.tar.gz", baseName, timestamp)
}

// uploadToMinio загружает файл в MinIO
func uploadToMinio(bucketName, objectName, filePath string) error {
	// Загрузка переменных окружения из .env файла
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("error loading .env file: %v", err)
	}

	// Чтение переменных окружения
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := true

	// Инициализация MinIO клиента
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return fmt.Errorf("error initializing minio client: %v", err)
	}

	// Загрузка файла
	ctx := context.Background()
	_, err = minioClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("error uploading file: %v", err)
	}

	return nil
}

func main() {
	// Чтение конфигурации из YAML файла
	file, err := os.Open("config.yml")
	if err != nil {
		log.Fatalf("Failed to open config file: %v", err)
	}
	defer file.Close()

	var config BackupConfig
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		log.Fatalf("Failed to decode config file: %v", err)
	}

	// Получение имени бакета из переменных окружения
	bucketName := os.Getenv("MINIO_BUCKET_NAME")
	if bucketName == "" {
		log.Fatalf("MINIO_BUCKET_NAME is not set in .env file")
	}

	for _, backup := range config.Backups {
		if backup.Type != "folder" && backup.Type != "volume" {
			log.Printf("Unsupported backup type: %s", backup.Type)
			continue
		}

		// Генерация имени архива с текущей датой и временем
		tarName := generateTarName(backup.Name)
		tmpFilePath := filepath.Join("/tmp", tarName)

		var err error

		if backup.Type == "folder" {
			_, err = tarFolder(backup.Source, tmpFilePath)
		} else if backup.Type == "volume" {
			err = createDockerVolumeBackup(backup.Source, tmpFilePath)
		}

		if err != nil {
			log.Printf("Failed to create backup for %s: %v", backup.Source, err)
			continue
		}

		// Определение пути в бакете, если указан path-save
		objectName := tarName
		if backup.PathSave != "" {
			objectName = filepath.Join(backup.PathSave, tarName)
		}

		// Загрузка файла в MinIO
		err = uploadToMinio(bucketName, objectName, tmpFilePath)
		if err != nil {
			log.Printf("Failed to upload %s to MinIO: %v", objectName, err)
			continue
		}

		// Удаление временного файла
		err = os.Remove(tmpFilePath)
		if err != nil {
			log.Printf("Failed to remove temporary file %s: %v", tmpFilePath, err)
		} else {
			fmt.Printf("Temporary file %s removed\n", tmpFilePath)
		}

		fmt.Printf("File %s successfully uploaded to MinIO\n", objectName)
	}
}
