package zaplogger

import (
	"go.uber.org/zap"

	"github.com/eugene982/url-shortener/internal/logger"
)

// Описание структуры логгера
type ZapLogger struct {
	zap *zap.Logger
}

// Конструктор нового логгера
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	var cfg zap.Config
	if lvl.Level() == zap.DebugLevel {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}

	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	logger.Log = &ZapLogger{zl}
	return nil
}

// Отладочные сообщения
func (z *ZapLogger) Debug(msg string, a ...any) {
	z.zap.Sugar().Debugw(msg, a...)
}

// Информационные сообщения
func (z *ZapLogger) Info(msg string, a ...any) {
	z.zap.Sugar().Infow(msg, a...)
}

// Предупреждения
func (z *ZapLogger) Warn(msg string, a ...any) {
	z.zap.Sugar().Warnw(msg, a...)
}

// Ошибки
func (z *ZapLogger) Error(err error, a ...any) {
	z.zap.Sugar().Errorw(err.Error(), a...)
}
