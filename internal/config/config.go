package config

import (
	"flag"
	"os"
	"time"

	"github.com/caarlos0/env/v8"
	"github.com/eugene982/url-shortener/internal/logger"
)

// Configuration структура получения данных из командной строки и окружения.
type Configuration struct {
	ServAddr        string        `env:"SERVER_ADDRESS"` // адрес сервера
	BaseURL         string        `env:"BASE_URL"`       // базовый адрес
	Timeout         time.Duration `env:"SERVER_TIMEOUT"`
	LogLevel        string        `env:"LOG_LEVEL"` // уровень логирования
	FileStoragePath string        `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string        `env:"DATABASE_DSN"`
	ProfAddr        string        `env:"PPROF_ADDRESS"`
	EnableHTTPS     bool          `env:"ENABLE_HTTPS"`
}

var config Configuration

// Config возвращаем копию конфигурации полученную из флагов и окружения.
func Config() Configuration {
	// устанавливаем переменные для флага по умолчанию
	flag.StringVar(&config.ServAddr, "a", ":8080", "server address")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "base address")
	flag.DurationVar(&config.Timeout, "t", 30*time.Second, "server timeout")
	flag.StringVar(&config.LogLevel, "l", "info", "log level")
	flag.StringVar(&config.FileStoragePath, "f", "/tmp/short-url-db.json", "file storage path")

	flag.StringVar(&config.DatabaseDSN, "d", "", "postgres connection string")
	//flag.StringVar(&config.DatabaseDSN, "d", "postgres://test:test@localhost/url_shorten", "postgres connection string")

	flag.StringVar(&config.ProfAddr, "p", ":8081", "pprof server address")
	flag.BoolVar(&config.EnableHTTPS, "s", false, "enable HTTPS")

	// получаем конфигурацию из флагов и/или окружения
	flag.Parse()
	if err := env.Parse(&config); err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	return config
}
