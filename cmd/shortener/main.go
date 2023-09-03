// Сервис сокращения ссылок.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/eugene982/url-shortener/internal/app"
	"github.com/eugene982/url-shortener/internal/config"
	"github.com/eugene982/url-shortener/internal/shortener"
	"github.com/jmoiron/sqlx"

	"github.com/eugene982/url-shortener/internal/storage"
	"github.com/eugene982/url-shortener/internal/storage/memstore"
	"github.com/eugene982/url-shortener/internal/storage/pgxstore"

	"github.com/eugene982/url-shortener/internal/logger"
	"github.com/eugene982/url-shortener/internal/logger/zaplogger"
)

const (
	// сколько ждём времени на корректное завершение работы сервера
	closeServerTimeout = time.Second * 3
)

func main() {

	if err := run(); err != nil {
		log.Fatal(err)
	}

}

// Установка параметров сервера и его запуск
func run() error {

	// захват прерывания процесса
	ctxInterrupt, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	conf := config.Config()

	err := zaplogger.Initialize(conf.LogLevel)
	if err != nil {
		return err
	}

	var (
		store storage.Storage
		db    *sqlx.DB
	)

	if conf.DatabaseDSN != "" {
		//"postgres://username:password@localhost:5432/database_name"
		db, err = sqlx.Open("pgx", conf.DatabaseDSN)
		if err != nil {
			logger.Error(err)
			return err
		}
		store, err = pgxstore.New(db)
		logger.Info("new pgxstore", "dsn", conf.DatabaseDSN)
	} else {
		store, err = memstore.New(conf.FileStoragePath)
		logger.Info("new memstore", "file", conf.FileStoragePath)
	}

	if err != nil {
		return err
	}

	sh := shortener.NewSimpleShortener()

	application, err := app.NewApplication(sh, store, conf.BaseURL)
	if err != nil {
		logger.Error(err)
		return err
	}

	// Установим таймауты, вдруг соединение будет нестабильным
	s := &http.Server{
		ReadTimeout:  time.Duration(conf.Timeout) * time.Second,
		WriteTimeout: time.Duration(conf.Timeout) * time.Second,
		Addr:         conf.ServAddr,
		Handler:      app.NewRouter(application),
	}

	logger.Info("service start",
		"config", conf,
	)

	// запуск сервера в горутине
	srvErr := make(chan error)
	go func() {
		srvErr <- s.ListenAndServe()
	}()

	// ждём что раньше случится, ошибка старта сервера
	// или пользователь прервёт программу
	select {
	case <-ctxInterrupt.Done():
		// прервано пользователем
	case err := <-srvErr:
		// сервер не смог стартануть, некорректый адрес, занят порт...
		// эту ошибку логируем отдельно. В любом случае, нужно освободить ресурсы
		logger.Error(fmt.Errorf("error start server: %w", err))
	}

	// стартуем завершение сервера
	closeErr := make(chan error)
	go func() {
		closeErr <- application.Close()
	}()

	// Ждём пока сервер сам завершится
	// или за отведённое время
	ctxTimeout, stop := context.WithTimeout(context.Background(), closeServerTimeout)
	defer stop()

	select {
	case <-ctxTimeout.Done():
		logger.Warn("stop server on timeout")
		return nil
	case err := <-closeErr:
		if err != nil {
			logger.Error(fmt.Errorf("application close server: %w", err))
		}
		logger.Info("stop server gracefull")
		return err
	}
}
