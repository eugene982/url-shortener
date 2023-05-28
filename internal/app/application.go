// Пакет приложения
package app

import (
	"fmt"
	"strings"

	"github.com/eugene982/url-shortener/internal/filestorage"
)

// Сокращатель ссылок
type Shortener interface {
	Short(string) string
}

// Хранитель ссылок
type Storage interface {
	GetAddr(short string) (addr string, ok bool)
	Set(addr string, short string)
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
	shortener   Shortener
	store       Storage
	logger      Logger
	baseURL     string
	fileStorage *filestorage.FileStorage
}

// Функция конструктор приложения.
func NewApplication(shortener Shortener, store Storage, logger Logger,
	baseURL string, fileSorePath string) (*Application, error) {

	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	var (
		err         error
		fileStorage *filestorage.FileStorage
	)

	// хранение ранее созданных сокращений в файле
	// для восстановления после перезапуска.
	if fileSorePath != "" {
		if fileStorage, err = filestorage.New(fileSorePath); err != nil {
			logger.Error(fmt.Errorf("error open file storage: %w", err))
			return nil, err
		}

		urls, err := fileStorage.ReadAll()
		if err != nil {
			logger.Error(fmt.Errorf("error read from file storage: %w", err))
			return nil, err
		}
		// переносим все ранее сохранённые значения из файла
		for _, v := range urls {
			store.Set(v.OriginalURL, v.ShortURL)
		}
	}

	return &Application{
		shortener:   shortener,
		store:       store,
		logger:      logger,
		baseURL:     baseURL,
		fileStorage: fileStorage}, nil
}

// закрываем приложение
func (a *Application) Close() (err error) {
	err = a.fileStorage.Close()
	if err != nil {
		a.logger.Error(fmt.Errorf("error close file storage: %w", err))
	}
	return
}
