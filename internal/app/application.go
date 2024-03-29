// Package app приложение
package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/acme/autocert"

	"github.com/eugene982/url-shortener/internal/config"
	"github.com/eugene982/url-shortener/internal/logger"
	"github.com/eugene982/url-shortener/internal/shortener"
	"github.com/eugene982/url-shortener/internal/storage"
	"github.com/eugene982/url-shortener/internal/storage/memstore"
	"github.com/eugene982/url-shortener/internal/storage/pgxstore"
)

const (
	// размер буфферизированного кана по удалению ссылок
	delShortChanSize = 256
	delShortDuration = time.Second
)

// Application основное приложение
type Application struct {
	shortener     shortener.Shortener
	store         storage.Storage
	baseURL       string
	server        *http.Server
	profServer    *http.Server
	delShortChan  chan deleteUserData
	stopDelChan   chan struct{}
	trustedSubnet string
	grpcServer    *GRPCServer
}

func New(conf config.Configuration) (*Application, error) {
	var (
		app Application
		db  *sqlx.DB
		err error
	)

	app.baseURL = conf.BaseURL
	if !strings.HasSuffix(conf.BaseURL, "/") {
		app.baseURL += "/"
	}

	if conf.DatabaseDSN != "" {
		//"postgres://username:password@localhost:5432/database_name"
		db, err = sqlx.Open("pgx", conf.DatabaseDSN)
		if err != nil {
			return nil, fmt.Errorf("error open sql dartabase: %w", err)
		}
		if app.store, err = pgxstore.New(db); err != nil {
			return nil, fmt.Errorf("error create postgres store: %w", err)
		}
		logger.Info("new pgxstore", "dsn", conf.DatabaseDSN)

	} else {
		if app.store, err = memstore.New(conf.FileStoragePath); err != nil {
			return nil, fmt.Errorf("error create mem store: %w", err)
		}
		logger.Info("new memstore", "file", conf.FileStoragePath)
	}

	app.trustedSubnet = conf.TrustedSubnet
	app.shortener = shortener.NewSimpleShortener()

	app.stopDelChan = make(chan struct{})
	app.delShortChan = make(chan deleteUserData, delShortChanSize)

	// Установим таймауты, вдруг соединение будет нестабильным
	app.server = &http.Server{
		ReadTimeout:  conf.Timeout,
		WriteTimeout: conf.Timeout,
		Addr:         conf.ServAddr,
		Handler:      NewRouter(&app),
	}

	if conf.EnableHTTPS {
		// конструируем менеджер TLS-сертификатов
		manager := &autocert.Manager{
			// директория хранения сертификатов
			Cache: autocert.DirCache("cache-dir"),
			// функция, принимающая Terms of Service издателя сертификатов
			Prompt: autocert.AcceptTOS,
			// перечень документов, для которых будут поддерживаться сертификаты
			HostPolicy: autocert.HostWhitelist("shortener.ru", "www.shortener.ru"),
		}
		app.server.TLSConfig = manager.TLSConfig()
	}

	// Установим сервер сбора отладочной информации
	app.profServer = &http.Server{
		ReadTimeout:  conf.Timeout,
		WriteTimeout: conf.Timeout,
		Addr:         conf.ProfAddr,
		Handler:      newProfRouter(),
	}

	// Настраиваем gRPC-сервер
	app.grpcServer, err = NewGRPCServer(&app, conf.GRPCAddr)
	if err != nil {
		return nil, err
	}

	return &app, nil
}

// Start - запуск сервера.
// Запуск прослушивания канала на удаление ссылок
func (a *Application) Start() error {
	go a.startDeletionShortUrls()
	go func() {
		err := a.profServer.ListenAndServe()
		if err != nil {
			logger.Error(fmt.Errorf("error start pprof server: %w", err))
		}
	}()

	go func() {
		err := a.grpcServer.Start()
		if err != nil {
			logger.Error(fmt.Errorf("error start gRPC server: %w", err))
		}
	}()

	if a.server.TLSConfig != nil {
		return a.server.ListenAndServeTLS("", "")
	}
	return a.server.ListenAndServe()
}

// Stop закрываем приложение.
func (a *Application) Stop() (err error) {
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

// Структура для складывания в канал пары Пользоватьль - Ссылки
type deleteUserData struct {
	userID    string
	shortURLs []string
}

// DeleteUserShortAsync - запуск асинхронного удаления ссылок пользователя.
// Добавляем в канал список ссылок к удалению для указанного пользователя
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
