package backup

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/JamesStewy/go-mysqldump"
	_ "github.com/go-sql-driver/mysql"
)

// extractDBName извлекает имя базы данных из строки подключения
func extractDBName(connString string) string {
	// Найдите последний символ '/' и верните все, что после него
	lastSlash := strings.LastIndex(connString, "/")
	if lastSlash == -1 || lastSlash == len(connString)-1 {
		return ""
	}
	return connString[lastSlash+1:]
}

// BackupMySQL выполняет резервное копирование базы данных MySQL
func BackupMySQL(connString, outputDir string) (string, error) {
	// Получаем имя базы данных из строки подключения
	dbName := extractDBName(connString)
	if dbName == "" {
		return "", fmt.Errorf("database name is empty")
	}

	// Удаляем префикс "mysql://" из строки подключения, если он есть
	if strings.HasPrefix(connString, "mysql://") {
		connString = strings.TrimPrefix(connString, "mysql://")
	}

	filename := fmt.Sprintf("%s-%s", dbName, "2006-01-02T15:04:05Z")

	// Открытие подключения к базе данных
	db, err := sql.Open("mysql", connString)
	if err != nil {
		return "", fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	// Регистрация базы данных с mysqldump
	dumper, err := mysqldump.Register(db, outputDir, filename)
	if err != nil {
		return "", fmt.Errorf("error registering database: %w", err)
	}

	// Выполняем дамп базы данных
	dumpFilename, err := dumper.Dump()
	if err != nil {
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

	// Закрытие дампера
	dumper.Close()

	return archiveName, nil
}
