package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// ConfigBackup представляет один элемент конфигурации резервного копирования
type ConfigBackup struct {
	Name     string `yaml:"name"`                // Имя резервной копии
	Source   string `yaml:"source"`              // Источник данных для резервного копирования (папка или volume)
	Type     string `yaml:"type"`                // Тип данных ("folder" или "volume")
	PathSave string `yaml:"path-save,omitempty"` // Путь для сохранения в бакете (опционально)
	Schedule string `yaml:"schedule,omitempty"`  // Расписание для автоматического резервного копирования
}

// BackupConfig представляет полную конфигурацию резервного копирования
type BackupConfig struct {
	Backups []ConfigBackup `yaml:"backups"` // Список всех резервных копий
}

// LoadConfig загружает конфигурацию резервного копирования из YAML файла
func LoadConfig(filePath string) (*BackupConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config BackupConfig
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
