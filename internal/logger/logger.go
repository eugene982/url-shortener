package logger

import (
	"log"
)

// Logger интерфейс логгера
type Logger interface {
	Debug(msg string, pair ...any)
	Info(msg string, pair ...any)
	Warn(msg string, pair ...any)
	Error(err error, pair ...any)
}

var Log Logger

// Debug запись сообщения отладки.
func Debug(msg string, pair ...any) {
	if Log == nil {
		log.Println("DEBUG", msg, pair)
		return
	}
	Log.Debug(msg, pair...)
}

// Info запись информационного сообщения.
func Info(msg string, pair ...any) {
	if Log == nil {
		log.Println("INFO", msg, pair)
		return
	}
	Log.Info(msg, pair...)
}

// Warn запись предупреждения
func Warn(msg string, pair ...any) {
	if Log == nil {
		log.Println("WARN", msg, pair)
		return
	}
	Log.Warn(msg, pair...)
}

// Error запись об ощибке.
func Error(err error, pair ...any) {
	if Log == nil {
		log.Println("Error", err, pair)
		return
	}
	Log.Error(err, pair...)
}
