package config

import (
	"encoding/json"
	"flag"
	"os"
	"time"

	"github.com/caarlos0/env/v8"
)

// Configuration структура получения данных из командной строки и окружения.
type Configuration struct {
	ServAddr        string        `env:"SERVER_ADDRESS"` // адрес сервера
	ProfAddr        string        `env:"PPROF_ADDRESS"`
	GRPCAddr        string        `env:"GRPC_ADDRESS"`
	BaseURL         string        `env:"BASE_URL"` // базовый адрес
	Timeout         time.Duration `env:"SERVER_TIMEOUT"`
	LogLevel        string        `env:"LOG_LEVEL"` // уровень логирования
	FileStoragePath string        `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string        `env:"DATABASE_DSN"`
	EnableHTTPS     bool          `env:"ENABLE_HTTPS"`
	ConfigFile      string        `env:"CONFIG"`
	TrustedSubnet   string        `env:"TRUSTED_SUBNET"`
}

// JSONConfiguration структура файла конфигурации
type JSONConfiguration struct {
	ServAddr        *string `json:"server_address,omitempty"`
	BaseURL         *string `json:"base_url,omitempty"`
	FileStoragePath *string `json:"file_storage_path,omitempty"`
	DatabaseDSN     *string `json:"database_dsn,omitempty"`
	EnableHTTPS     *bool   `json:"enable_https,omitempty"`
	TrustedSubnet   *string `json:"trusted_subnet,omitempty"`
}

var config Configuration

// Config возвращаем копию конфигурации полученную из флагов и окружения.
func Config() (Configuration, error) {

	// устанавливаем переменные для флага по умолчанию
	flag.StringVar(&config.ServAddr, "a", ":8080", "server address")
	flag.StringVar(&config.ProfAddr, "p", ":8081", "pprof server address")
	flag.StringVar(&config.GRPCAddr, "g", ":8082", "grpc server address")

	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "base address")
	flag.DurationVar(&config.Timeout, "o", 30*time.Second, "server timeout")
	flag.StringVar(&config.LogLevel, "l", "info", "log level")
	flag.StringVar(&config.FileStoragePath, "f", "/tmp/short-url-db.json", "file storage path")

	flag.StringVar(&config.DatabaseDSN, "d", "", "postgres connection string")
	//flag.StringVar(&config.DatabaseDSN, "d", "postgres://test:test@localhost/url_shorten", "postgres connection string")

	flag.BoolVar(&config.EnableHTTPS, "s", false, "enable HTTPS")
	flag.StringVar(&config.TrustedSubnet, "t", "", "trusted subnet")

	// файл конфигурации
	flag.StringVar(&config.ConfigFile, "c", "", "json config file")

	// получаем конфигурацию из флагов
	flag.Parse()

	// поищем путь и в переменных окружения
	if config.ConfigFile != "" {
		if err := decodeJsonConfigFile(config.ConfigFile); err != nil {
			return Configuration{}, err
		}
	}

	// получаем конфигурацию из окружения
	if err := env.Parse(&config); err != nil {
		return Configuration{}, err
	}
	return config, nil
}

func decodeJsonConfigFile(fname string) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}

	var conf JSONConfiguration
	if err := json.NewDecoder(f).Decode(&conf); err != nil {
		return err
	}

	// Проверка наличия флага
	// Необходимо исключить флаги установленные по умолчанию
	// и имеющиеся в файле
	reserve := make(map[string]bool, 6)
	flag.Visit(func(f *flag.Flag) {
		reserve[f.Name] = true
	})

	if conf.ServAddr != nil && !reserve["a"] {
		config.ServAddr = *conf.ServAddr
	}
	if conf.BaseURL != nil && !reserve["b"] {
		config.BaseURL = *conf.BaseURL
	}
	if conf.FileStoragePath != nil && !reserve["f"] {
		config.FileStoragePath = *conf.FileStoragePath
	}
	if conf.DatabaseDSN != nil && !reserve["d"] {
		config.DatabaseDSN = *conf.DatabaseDSN
	}
	if conf.EnableHTTPS != nil && !reserve["s"] {
		config.EnableHTTPS = *conf.EnableHTTPS
	}
	if conf.TrustedSubnet != nil && !reserve["t"] {
		config.TrustedSubnet = *conf.TrustedSubnet
	}
	return nil
}
