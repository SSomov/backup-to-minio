package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/JCoupalK/go-pgdump"
)

// BackupPostgres выполняет резервное копирование базы данных PostgreSQL
func BackupPostgres(connString, outputDir string) (string, error) {
	currentTime := time.Now()
	dumpFilename := filepath.Join(outputDir, fmt.Sprintf("backup-%s.sql", currentTime.Format("20060102T150405")))

	// Создаем новый экземпляр дампера
	dumper := pgdump.NewDumper(connString, 50)

	// Выполняем дамп базы данных
	if err := dumper.DumpDatabase(dumpFilename, &pgdump.TableOptions{
		Schema: "public",
	}); err != nil {
		return "", fmt.Errorf("error dumping database: %w", err)
	}

	// Сжимаем файл дампа
	archiveName := GenerateTarName("backup") // Имя архива можно изменить на любое
	archivePath := filepath.Join(outputDir, archiveName)

	if _, err := TarFolder(dumpFilename, archivePath); err != nil {
		return "", fmt.Errorf("error compressing dump file: %w", err)
	}

	// Удаляем оригинальный файл дампа
	if err := os.Remove(dumpFilename); err != nil {
		fmt.Printf("Failed to remove original dump file %s: %v\n", dumpFilename, err)
	} else {
		fmt.Printf("Original dump file %s removed\n", dumpFilename)
	}

	return archivePath, nil
}
