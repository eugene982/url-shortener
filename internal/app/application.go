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
	GetShort(string) (string, bool)
	Set(string, string) bool
}

// Управлятель ссылок
type Application struct {
	shortener Shortener
	store     Storage
	baseUrl   string
}

// Функция конструктор приложения.
func NewApplication(shortener Shortener, store Storage, baseUrl string) *Application {
	if baseUrl != "" && !strings.HasSuffix(baseUrl, "/") {
		baseUrl += "/"
	}
	return &Application{shortener, store, baseUrl}
}
