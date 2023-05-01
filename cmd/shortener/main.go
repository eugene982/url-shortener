// Сервис сокращения ссылок.
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/eugene982/url-shortener/internal/app"
)

var address = "localhost:8080"
var timeout = 30 * time.Second

func main() {

	if err := run(); err != nil {
		log.Fatal(err)
	}

}

// Установка параметров сервера и его запуск
func run() error {

	sh := app.NewSimpleShortener()
	st := app.NewMemstore()

	application := app.NewApplication(sh, st)

	// Установим таймауты, вдруг соединение будет нестабильным
	s := &http.Server{
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		Addr:         address,
		Handler:      application.GetMux(),
	}

	return s.ListenAndServe()
}
