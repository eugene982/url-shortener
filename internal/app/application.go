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
	delShortChan chan deleteUserData
	stopDelChan  chan struct{}
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
		delShortChan: make(chan deleteUserData, delShortChanSize),
	}

	app.stopDelChan = make(chan struct{})
	go app.startDeletionShortUrls()

	return app, nil
}

// закрываем приложение
func (a *Application) Close() (err error) {
	a.stopDelChan <- struct{}{}
	if err = a.store.Close(); err != nil {
		logger.Error(err)
	}
	return
}

// Обработка очереди пометки на удаление
func (a *Application) startDeletionShortUrls() {

	ticker := time.NewTicker(delShortDuration)
	delete := make([]deleteUserData, 0)

	// копим пачку ссылок
	for {
		select {
		case <-a.stopDelChan:
			return // завершаем горутину
		case d := <-a.delShortChan:
			delete = append(delete, d)

		case <-ticker.C:
			if len(delete) == 0 {
				continue
			}

			ctx := context.Background()

			// сгруппируем по пользователю
			usersURLs := map[string][]string{}
			for _, d := range delete {
				usersURLs[d.userID] = append(usersURLs[d.userID], d.shortURLs...)
			}

			// Удалим все ссылки всех пользователей разом.
			delShortURLs := make([]string, 0)

			// по каждому пользователю получим список ссылок
			// и выберем только те что есть в хранилище
			for userID, shortURLs := range usersURLs {

				data, err := a.store.GetUserURLs(ctx, userID)
				if err != nil {
					logger.Error(err)
					break // при ошибке выходим и
				}

				for _, d := range data {
					for _, s := range shortURLs {
						if d.ShortURL == s {
							delShortURLs = append(delShortURLs, s)
						}
					}
				}
			}

			err := a.store.DeleteShort(ctx, delShortURLs)
			if err != nil {
				logger.Error(err)
				continue //
			}
			delete = delete[:0] // очищаем при успешном удалении
		}
	}
}

func (a *Application) GetBaseURL() string {
	return a.baseURL
}

// Структура для складывания в канал пары Пользоватьль - Ссылки
type deleteUserData struct {
	userID    string
	shortURLs []string
}

// добавляем в канал список ссылок к удалению для указанного пользователя
func (a Application) DeleteUserShortAsync(userID string, shorts []string) {

	// добавляем все данные без разбора.
	// Проверять принадлежность ссылки пользователю будем асинхронно в горутине
	if len(shorts) > 0 {
		a.delShortChan <- deleteUserData{
			userID:    userID,
			shortURLs: shorts,
		}
	}

}
