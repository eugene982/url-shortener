package logger

import (
	"log"
)

// Логгер
type Logger interface {
	Debug(msg string, pair ...any)
	Info(msg string, pair ...any)
	Warn(msg string, pair ...any)
	Error(err error, pair ...any)
}

var Log Logger

func Debug(msg string, pair ...any) {
	if Log == nil {
		log.Println("DEBUG", msg, pair)
		return
	}
	Log.Debug(msg, pair...)
}

func Info(msg string, pair ...any) {
	if Log == nil {
		log.Println("INFO", msg, pair)
		return
	}
	Log.Info(msg, pair...)
}

func Warn(msg string, pair ...any) {
	if Log == nil {
		log.Println("WARN", msg, pair)
		return
	}
	Log.Warn(msg, pair...)
}

func Error(err error, pair ...any) {
	if Log == nil {
		log.Println("Error", err, pair)
		return
	}
	Log.Error(err, pair...)
}
