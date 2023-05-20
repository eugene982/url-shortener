// Пакет приложения
package app

import (
	"strings"
)

// Сокращатель ссылок
type Shortener interface {
	Short(string) string
}

// Хранитель ссылок
type Storage interface {
	GetAddr(string) (string, bool)
	Set(string, string)
}

// Логгер
type Logger interface {
	Debug(msg string, pair ...any)
	Info(msg string, pair ...any)
	Warn(msg string, pair ...any)
	Error(err error, pair ...any)
}

// Управлятель ссылок
type Application struct {
	shortener Shortener
	store     Storage
	logger    Logger
	baseURL   string
}

// Функция конструктор приложения.
func NewApplication(shortener Shortener, store Storage, logger Logger, baseURL string) *Application {
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	return &Application{shortener, store, logger, baseURL}
}
