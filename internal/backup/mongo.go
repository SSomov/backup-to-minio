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

// MongoDBParams содержит параметры подключения к MongoDB
type MongoDBParams struct {
	URI        string
	Host       string
	Port       string
	DBName     string
	Username   string
	Password   string
	AuthDB     string
}

// parseMongoConnString парсит строку подключения MongoDB
func parseMongoConnString(connString string) (*MongoDBParams, error) {
	// Удаляем префикс mongodb:// при наличии
	connString = strings.TrimPrefix(connString, "mongodb://")

	u, err := url.Parse(connString)
	if err != nil {
		return nil, fmt.Errorf("invalid MongoDB connection string: %w", err)
	}

	password, _ := u.User.Password()
	host, port := u.Hostname(), u.Port()
	if port == "" {
		port = "27017"
	}

	dbName := strings.TrimPrefix(u.Path, "/")
	if strings.Contains(dbName, "/") {
		dbName = strings.Split(dbName, "/")[0]
	}

	query := u.Query()
	authDB := query.Get("authSource")

	return &MongoDBParams{
		URI:      connString,
		Host:     host,
		Port:     port,
		DBName:   dbName,
		Username: u.User.Username(),
		Password: password,
		AuthDB:   authDB,
	}, nil
}

// BackupMongoDB выполняет резервное копирование MongoDB
func BackupMongoDB(connString, outputDir string) (string, error) {
	params, err := parseMongoConnString(connString)
	if err != nil {
		return "", fmt.Errorf("MongoDB connection error: %w", err)
	}

	if params.DBName == "" {
		return "", fmt.Errorf("database name is required")
	}

	// Генерируем имя файла
	timestamp := time.Now().Format("2006-01-02T15-04-05Z")
	dumpDir := filepath.Join(outputDir, fmt.Sprintf("%s-%s", params.DBName, timestamp))
	archiveName := dumpDir + ".gz"

	// Формируем команду mongodump
	cmdArgs := []string{
		"--host", params.Host + ":" + params.Port,
		"--db", params.DBName,
		"--out", dumpDir,
	}

	if params.Username != "" {
		cmdArgs = append(cmdArgs, "--username", params.Username)
	}
	if params.Password != "" {
		cmdArgs = append(cmdArgs, "--password", params.Password)
	}
	if params.AuthDB != "" {
		cmdArgs = append(cmdArgs, "--authenticationDatabase", params.AuthDB)
	}

	cmd := exec.Command("mongodump", cmdArgs...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Выполняем команду
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("mongodump failed: %v\nError: %s", err, stderr.String())
	}

	// Сжимаем результат
	if err := GzCompressDir(dumpDir, archiveName); err != nil {
		return "", fmt.Errorf("compression failed: %w", err)
	}

	// Удаляем временную директорию
	if err := os.RemoveAll(dumpDir); err != nil {
		fmt.Printf("Warning: failed to remove temp directory %s: %v\n", dumpDir, err)
	}

	return archiveName, nil
}

// GzCompressDir сжимает директорию в .tar.gz
func GzCompressDir(source, target string) error {
	cmd := exec.Command("tar", "-czf", target, "-C", filepath.Dir(source), filepath.Base(source))
	
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tar compression failed: %v\nError: %s", err, stderr.String())
	}
	return nil
}