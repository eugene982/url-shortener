package config

import (
	"flag"

	"github.com/caarlos0/env/v8"
)

// Объявление структуры конфигурации
type Configuration struct {
	ServAddr        string `env:"SERVER_ADDRESS"` // адрес сервера
	BaseURL         string `env:"BASE_URL"`       // базовый адрес
	Timeout         int
	LogLevel        string `env:"LOG_LEVEL"` // уровень логирования
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

var config Configuration

func init() {
	// устанавливаем переменные для флага по умолчанию
	flag.StringVar(&config.ServAddr, "a", ":8080", "server address")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "base address")
	flag.IntVar(&config.Timeout, "t", 30, "timeout in seconds")
	flag.StringVar(&config.LogLevel, "l", "info", "log level")
	flag.StringVar(&config.FileStoragePath, "f", "/tmp/short-url-db.json", "file storage path")
	flag.StringVar(&config.DatabaseDSN, "d", "", "postgres connection string")

	// получаем конфигурацию из флагов и/или окружения
	flag.Parse()
	env.Parse(&config)
}

// Возвращаем копию конфигурации полученную из флагов и окружения
func Config() Configuration {
	return config
}
