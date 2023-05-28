// Сервис сокращения ссылок.
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/eugene982/url-shortener/internal/app"
	"github.com/eugene982/url-shortener/internal/config"
	"github.com/eugene982/url-shortener/internal/logger"
	"github.com/eugene982/url-shortener/internal/shortener"
	"github.com/eugene982/url-shortener/internal/storage"
)

func main() {

	if err := run(); err != nil {
		log.Fatal(err)
	}

}

// Установка параметров сервера и его запуск
func run() error {

	conf := config.Config()

	sh := shortener.NewSimpleShortener()
	st := storage.NewMemstore()
	logger, err := logger.NewZapLogger(conf.LogLevel)
	if err != nil {
		return err
	}

	application, err := app.NewApplication(sh, st, logger, conf.BaseURL, conf.FileStoragePeth)
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
