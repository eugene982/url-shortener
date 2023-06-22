// Сервис сокращения ссылок.
package main

import (
	"log"
	"net/http"
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

func main() {

	if err := run(); err != nil {
		log.Fatal(err)
	}

}

// Установка параметров сервера и его запуск
func run() error {

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
		return err
	}
	defer application.Close()

	// Установим таймауты, вдруг соединение будет нестабильным
	s := &http.Server{
		ReadTimeout:  time.Duration(conf.Timeout) * time.Second,
		WriteTimeout: time.Duration(conf.Timeout) * time.Second,
		Addr:         conf.ServAddr,
		Handler:      application.NewRouter(),
	}

	logger.Info("service start",
		"addres", conf.ServAddr,
		"base_url", conf.BaseURL,
	)

	err = s.ListenAndServe()
	logger.Info("service stop",
		"error", err)
	return err
}
