// Сервис сокращения ссылок.
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/eugene982/url-shortener/internal/app"
	"github.com/eugene982/url-shortener/internal/config"
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

	sh := shortener.NewSimpleShortener()
	st := storage.NewMemstore()

	conf := config.GetConfig()
	application := app.NewApplication(sh, st, conf.BaseURL)

	// Установим таймауты, вдруг соединение будет нестабильным
	s := &http.Server{
		ReadTimeout:  time.Duration(conf.Timeout) * time.Second,
		WriteTimeout: time.Duration(conf.Timeout) * time.Second,
		Addr:         conf.ServAddr,
		Handler:      application.NewRouter(),
	}

	log.Println("service start on address:", conf.ServAddr)

	return s.ListenAndServe()
}
