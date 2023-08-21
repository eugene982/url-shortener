package zaplogger

import (
	"go.uber.org/zap"

	"github.com/eugene982/url-shortener/internal/logger"
)

// ZapLogger - логгер zap
type ZapLogger struct {
	zap *zap.Logger
}

// Initialize создание нового логгера.
// Сохранение ссылки в глобальную переменную Log.
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

// Debug Отладочные сообщения
func (z *ZapLogger) Debug(msg string, a ...any) {
	z.zap.Sugar().Debugw(msg, a...)
}

// Info Информационные сообщения
func (z *ZapLogger) Info(msg string, a ...any) {
	z.zap.Sugar().Infow(msg, a...)
}

// Warn Предупреждения
func (z *ZapLogger) Warn(msg string, a ...any) {
	z.zap.Sugar().Warnw(msg, a...)
}

// Error Ошибки
func (z *ZapLogger) Error(err error, a ...any) {
	z.zap.Sugar().Errorw(err.Error(), a...)
}
