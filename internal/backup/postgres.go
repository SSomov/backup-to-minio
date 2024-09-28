package backup

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/JCoupalK/go-pgdump"
)

// getDBNameFromConnString извлекает имя базы данных из строки подключения
func getDBNameFromConnString(connString string) (string, error) {
	u, err := url.Parse(connString)
	if err != nil {
		return "", fmt.Errorf("failed to parse connection string: %w", err)
	}

	dbName := u.Path
	if dbName == "" || dbName == "/" {
		return "", fmt.Errorf("database name not found in connection string")
	}

	return dbName[1:], nil // убираем ведущий слэш
}

// BackupPostgres выполняет резервное копирование базы данных PostgreSQL
func BackupPostgres(connString, outputDir string) (string, error) {
	dbName, err := getDBNameFromConnString(connString)
	if err != nil {
		return "", fmt.Errorf("failed to get database name: %w", err)
	}
	currentTime := time.Now()
	dumpFilename := filepath.Join(outputDir, fmt.Sprintf("%s-%s.sql", dbName, currentTime.Format("2006-01-02T15:04:05Z")))

	// Создаем новый экземпляр дампера
	dumper := pgdump.NewDumper(connString, 50)

	// Выполняем дамп базы данных
	if err := dumper.DumpDatabase(dumpFilename, &pgdump.TableOptions{
		Schema: "public",
	}); err != nil {
		return "", fmt.Errorf("error dumping database: %w", err)
	}

	archiveName, err := GzFile(dumpFilename)
	if err != nil {
		return "", fmt.Errorf("error compressing dump file: %w", err)
	}

	// Удаляем оригинальный файл дампа
	if err := os.Remove(dumpFilename); err != nil {
		fmt.Printf("Failed to remove original dump file %s: %v\n", dumpFilename, err)
	} else {
		fmt.Printf("Original dump file %s removed\n", dumpFilename)
	}

	return archiveName, nil
}
