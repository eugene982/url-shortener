package logger

import (
	"net/http"

	"go.uber.org/zap"
)

// Описание структуры логгера
type ZapLogger struct {
	zap *zap.Logger
}

// Конструктор нового логгера
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

// структура захвата Ответа сервиса для логирования
type LogResponseWriter struct {
	http.ResponseWriter
	size       int
	statusCode int
}

func (l *LogResponseWriter) Size() int       { return l.size }
func (l *LogResponseWriter) StatusCode() int { return l.statusCode }

// Проверка на тип, на всяки....
var _ http.ResponseWriter = (*LogResponseWriter)(nil)

func NewLogResponseWriter(r http.ResponseWriter) *LogResponseWriter {
	return &LogResponseWriter{r, 0, 0}
}

// Write implements http.ResponseWriter
func (l *LogResponseWriter) Write(bytes []byte) (size int, err error) {
	size, err = l.ResponseWriter.Write(bytes)
	l.size += size
	return
}

// WriteHeader implements http.ResponseWriter
func (l *LogResponseWriter) WriteHeader(statusCode int) {
	l.ResponseWriter.WriteHeader(statusCode)
	l.statusCode = statusCode
}
