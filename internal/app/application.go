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

// Управлятель ссылок
type Application struct {
	shortener Shortener
	store     Storage
	baseURL   string
}

// Функция конструктор приложения.
func NewApplication(shortener Shortener, store Storage, baseURL string) *Application {
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	return &Application{shortener, store, baseURL}
}
