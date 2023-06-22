// Пакет приложения
package app

import (
	"context"
	"strings"
	"time"

	"github.com/eugene982/url-shortener/internal/logger"
	"github.com/eugene982/url-shortener/internal/shortener"
	"github.com/eugene982/url-shortener/internal/storage"
)

const (
	// размер буфферизированного кана по удалению ссылок
	delShortChanSize = 256
	delShortDuration = time.Second
)

// Управлятель ссылок
type Application struct {
	shortener    shortener.Shortener
	store        storage.Storage
	baseURL      string
	delShortChan chan string
}

// Функция конструктор приложения.
func NewApplication(shortener shortener.Shortener,
	store storage.Storage, baseURL string) (*Application, error) {

	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	app := &Application{
		shortener:    shortener,
		store:        store,
		baseURL:      baseURL,
		delShortChan: make(chan string, delShortChanSize),
	}

	go app.deleteShortUrls()

	return app, nil
}

// закрываем приложение
func (a *Application) Close() (err error) {
	if err = a.store.Close(); err != nil {
		logger.Error(err)
	}
	return
}

// Обработка очереди пометки на удаление
func (a *Application) deleteShortUrls() {

	ticker := time.NewTicker(delShortDuration)
	delete := make([]string, 0)

	// копим пачку ссылок
	for {
		select {
		case short := <-a.delShortChan:
			delete = append(delete, short)
		case <-ticker.C:
			if len(delete) == 0 {
				continue
			}
			ctx, close := context.WithCancel(context.Background())

			err := a.store.DeleteShort(ctx, delete)
			if err != nil {
				logger.Error(err)
				continue //
			}

			delete = delete[:0] // очищаем при успешном удалении
			close()

		}
	}
}
