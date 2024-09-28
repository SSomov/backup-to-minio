package backup

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// TarFolder сжимает все файлы и папки в указанной директории и возвращает путь к созданному архиву.
func TarFolder(source, target string) (string, error) {
	// Создаем файл архива
	tarfile, err := os.Create(target)
	if err != nil {
		return "", fmt.Errorf("failed to create tar file: %w", err)
	}
	defer tarfile.Close()

	// Создаем gzip-архив
	gzw := gzip.NewWriter(tarfile)
	defer gzw.Close()

	// Создаем tar-архив
	tw := tar.NewWriter(gzw)
	defer tw.Close()

	baseDir := filepath.Base(source)

	// Добавляем корневую папку в архив
	if err := tw.WriteHeader(&tar.Header{
		Name:     baseDir + "/",
		Mode:     0755,
		Typeflag: tar.TypeDir,
	}); err != nil {
		return "", fmt.Errorf("failed to write header: %w", err)
	}

	// Проходим по всем файлам и папкам в указанной директории
	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Пропускаем директории, так как они будут обработаны рекурсивно
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		// Создаем заголовок для каждого файла
		header, err := tar.FileInfoHeader(info, relPath)
		if err != nil {
			return fmt.Errorf("failed to create tar header: %w", err)
		}

		// Устанавливаем имя файла внутри архива
		header.Name = filepath.Join(baseDir, relPath)

		// Записываем заголовок в архив
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}

		// Открываем файл для чтения
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		// Копируем содержимое файла в архив
		if _, err := io.Copy(tw, file); err != nil {
			return fmt.Errorf("failed to copy file content: %w", err)
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("error walking the path: %w", err)
	}

	return tarfile.Name(), nil
}

// GenerateTarName генерирует имя файла архива с текущей датой и временем в формате ISO 8601.
func GenerateTarName(baseName string) string {
	now := time.Now().UTC()
	timestamp := now.Format("2006-01-02T15:04:05Z")
	return fmt.Sprintf("%s_%s.tar.gz", baseName, timestamp)
}
