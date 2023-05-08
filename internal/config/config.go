package config

import "flag"

// структура конфигурации
type Config struct {
	Addr    string // адрес сервера
	Base    string // базовый адрес
	Timeout int
}

var flagConf Config

func init() {
	flag.StringVar(&flagConf.Addr, "a", ":8080", "server address")
	flag.StringVar(&flagConf.Base, "b", "", "base address") // http://localhost:8000/api
	flag.IntVar(&flagConf.Timeout, "t", 30, "timeout in seconds")
}

func GetConfig() Config {
	flag.Parse()
	return flagConf
}
