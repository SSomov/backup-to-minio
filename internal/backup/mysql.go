package backup

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// MySQLParams содержит параметры подключения к MySQL
type MySQLParams struct {
	User     string
	Password string
	Host     string
	Port     string
	DBName   string
}

// parseMySQLConnString парсит строку подключения MySQL
func parseMySQLConnString(connString string) (*MySQLParams, error) {
	// Удаляем префикс mysql:// при наличии
	connString = strings.TrimPrefix(connString, "mysql://")

	// Парсим как URL
	u, err := url.Parse(connString)
	if err != nil {
		return nil, fmt.Errorf("invalid connection string: %w", err)
	}

	// Извлекаем параметры
	password, _ := u.User.Password()
	host, port := u.Hostname(), u.Port()
	if port == "" {
		port = "3306" // Порт по умолчанию
	}

	return &MySQLParams{
		User:     u.User.Username(),
		Password: password,
		Host:     host,
		Port:     port,
		DBName:   strings.TrimPrefix(u.Path, "/"),
	}, nil
}

// BackupMySQL выполняет резервное копирование с использованием mysqldump
func BackupMySQL(connString, outputDir string) (string, error) {
	params, err := parseMySQLConnString(connString)
	if err != nil {
		return "", fmt.Errorf("connection string parse error: %w", err)
	}

	if params.DBName == "" {
		return "", fmt.Errorf("database name is required")
	}

	// Генерируем имя файла
	timestamp := time.Now().Format("2006-01-02T15-04-05Z")
	dumpFilename := filepath.Join(outputDir, fmt.Sprintf("%s-%s.sql", params.DBName, timestamp))

	// Формируем команду mysqldump
	cmd := exec.Command(
		"mysqldump",
		"-h", params.Host,
		"-P", params.Port,
		"-u", params.User,
		"--password="+params.Password,
		"--single-transaction",
		"--routines",
		"--triggers",
		params.DBName,
	)

	// Настраиваем вывод
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Создаем файл для дампа
	outputFile, err := os.Create(dumpFilename)
	if err != nil {
		return "", fmt.Errorf("failed to create dump file: %w", err)
	}
	defer outputFile.Close()
	cmd.Stdout = outputFile

	// Выполняем команду
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("mysqldump failed: %v\nStderr: %s", err, stderr.String())
	}

	// Сжимаем файл
	archiveName, err := GzFile(dumpFilename)
	if err != nil {
		return "", fmt.Errorf("compression failed: %w", err)
	}

	// Удаляем оригинальный файл
	if err := os.Remove(dumpFilename); err != nil {
		fmt.Printf("Warning: failed to remove temp file %s: %v\n", dumpFilename, err)
	}

	return archiveName, nil
}