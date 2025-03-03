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

// ConnectionParams содержит параметры подключения к PostgreSQL
type ConnectionParams struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// parseConnString парсит строку подключения PostgreSQL
func parseConnString(connString string) (*ConnectionParams, error) {
	if !strings.Contains(connString, "://") {
		connString = "postgres://" + connString
	}

	u, err := url.Parse(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	password, _ := u.User.Password()
	host, port := u.Hostname(), u.Port()
	if port == "" {
		port = "5432"
	}

	return &ConnectionParams{
		Host:     host,
		Port:     port,
		User:     u.User.Username(),
		Password: password,
		DBName:   strings.TrimPrefix(u.Path, "/"),
	}, nil
}

// BackupPostgres выполняет резервное копирование с использованием pg_basebackup
func BackupPostgres(connString, outputDir string) (string, error) {
	params, err := parseConnString(connString)
	if err != nil {
		return "", fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Создаем имя файла с временной меткой
	timestamp := time.Now().Format("2006-01-02T15-04-05Z")
	baseName := fmt.Sprintf("%s-%s.tar.gz", params.DBName, timestamp)
	dumpPath := filepath.Join(outputDir, baseName)

	// Формируем команду pg_basebackup
	cmd := exec.Command(
		"pg_basebackup",
		"-h", params.Host,
		"-p", params.Port,
		"-U", params.User,
		"-D", "-",    // Вывод в stdout
		"-Ft",        // Формат tar
		"-z",         // Сжатие
		"-P",         // Прогресс-бар
		"--label", fmt.Sprintf("%s_%s", params.DBName, timestamp),
	)

	// Настраиваем переменные окружения
	cmd.Env = append(os.Environ(),
		"PGPASSWORD="+params.Password,
		"PGDATABASE="+params.DBName,
	)

	// Перенаправляем stdout в файл
	outputFile, err := os.Create(dumpPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()
	cmd.Stdout = outputFile

	// Захватываем stderr для вывода ошибок
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Выполняем команду
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pg_basebackup failed: %v, stderr: %s", err, stderr.String())
	}

	return dumpPath, nil
}