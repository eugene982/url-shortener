package config

import (
	"flag"

	"github.com/caarlos0/env"
)

// Объявление структуры конфигурации
type Configuration struct {
	ServAddr string `env:"SERVER_ADDRESS"` // адрес сервера
	BaseURL  string `env:"BASE_URL"`       // базовый адрес
	Timeout  int
}

var config Configuration

func init() {
	// устанавливаем переменные для флага по умолчанию
	flag.StringVar(&config.ServAddr, "a", ":8080", "server address")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "base address")
	flag.IntVar(&config.Timeout, "t", 30, "timeout in seconds")

	// получаем конфигурацию из флагов и/или окружения
	flag.Parse()

	var envConf Configuration

	err := env.Parse(&envConf)
	if err == nil {
		if envConf.ServAddr != "" {
			config.ServAddr = envConf.ServAddr
		}
		if envConf.BaseURL != "" {
			config.BaseURL = envConf.BaseURL
		}
	}
}

// Возвращаем копию конфигурации полученную из флагов и окружения
func Config() Configuration {
	return config
}
