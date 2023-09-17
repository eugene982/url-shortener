// Сервис сокращения ссылок.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eugene982/url-shortener/internal/app"
	"github.com/eugene982/url-shortener/internal/config"

	"github.com/eugene982/url-shortener/internal/logger"
	"github.com/eugene982/url-shortener/internal/logger/zaplogger"
)

const (
	// сколько ждём времени на корректное завершение работы сервера
	closeServerTimeout = time.Second * 3
)

var (
	buildVersion string = "N/A" // версия сборки
	buildDate    string = "N/A" // дата сборки
	buildCommit  string = "N/A" // сомментарий сборки
)

func main() {

	fmt.Println("Build version:", buildVersion)
	fmt.Println("Build date:", buildDate)
	fmt.Println("Build commit:", buildCommit)

	if err := run(); err != nil {
		log.Fatal(err)
	}

}

// Установка параметров сервера и его запуск
func run() error {

	// захват прерывания процесса
	ctxInterrupt, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	conf := config.Config()

	err := zaplogger.Initialize(conf.LogLevel)
	if err != nil {
		return err
	}

	application, err := app.New(conf)
	if err != nil {
		return err
	}

	// запуск приложения в горутине
	srvErr := make(chan error)
	go func() {
		srvErr <- application.Start()
	}()

	logger.Info("service start",
		"config", conf,
	)

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
		closeErr <- application.Stop()
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
