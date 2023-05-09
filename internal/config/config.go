package config

import (
	"flag"

	"github.com/caarlos0/env"
)

// структура конфигурации
type Config struct {
	ServAddr string `env:"SERVER_ADDRESS"` // адрес сервера
	BaseURL  string `env:"BASE_URL"`       // базовый адрес
	Timeout  int
}

var flagConf Config

// устанавливаем переменные для флага по умолчанию
func init() {
	flag.StringVar(&flagConf.ServAddr, "a", ":8080", "server address")
	flag.StringVar(&flagConf.BaseURL, "b", "", "base address") // http://localhost:8000/api
	flag.IntVar(&flagConf.Timeout, "t", 30, "timeout in seconds")
}

// получаем конфигурацию из флагов и/или окружения
func GetConfig() Config {
	flag.Parse()

	var envConf Config
	err := env.Parse(&envConf)
	if err == nil {
		if envConf.ServAddr != "" {
			flagConf.ServAddr = envConf.ServAddr
		}
		if envConf.BaseURL != "" {
			flagConf.BaseURL = envConf.BaseURL
		}
	}

	return flagConf
}
