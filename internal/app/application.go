// Пакет приложения
package app

import (
	"strings"

	"github.com/eugene982/url-shortener/internal/logger"
	"github.com/eugene982/url-shortener/internal/shortener"
	"github.com/eugene982/url-shortener/internal/storage"
)

// Управлятель ссылок
type Application struct {
	shortener shortener.Shortener
	store     storage.Storage
	baseURL   string
}

// Функция конструктор приложения.
func NewApplication(shortener shortener.Shortener,
	store storage.Storage, baseURL string) (*Application, error) {

	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	return &Application{
		shortener: shortener,
		store:     store,
		baseURL:   baseURL,
	}, nil
}

// закрываем приложение
func (a *Application) Close() (err error) {
	if err = a.store.Close(); err != nil {
		logger.Error(err)
	}
	return
}
