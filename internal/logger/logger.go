package logger

import (
	"go.uber.org/zap"

	"github.com/eugene982/url-shortener/internal/app"
)

// Описание структуры логгера
type ZapLogger struct {
	zap *zap.Logger
}

// проверка на соответствие типа
var _ app.Logger = (*ZapLogger)(nil)

// Создание нового логгера
func NewZapLogger(level string) (*ZapLogger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}

	var cfg zap.Config
	if lvl.Level() == zap.DebugLevel {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}

	cfg.Level = lvl

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	return &ZapLogger{logger}, nil
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
