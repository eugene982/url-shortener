package app

import (
	"net/http"
	"time"
)

// Структура для временного хранения записываемых в ответ данных
type logResponseWriter struct {
	http.ResponseWriter
	size  int
	satus int
}

// Write implements http.ResponseWriter
func (l *logResponseWriter) Write(bytes []byte) (size int, err error) {
	size, err = l.ResponseWriter.Write(bytes)
	l.size += size
	return
}

// WriteHeader implements http.ResponseWriter
func (l *logResponseWriter) WriteHeader(statusCode int) {
	l.ResponseWriter.WriteHeader(statusCode)
	l.satus = statusCode
}

// Проверка на тип, на всяки....
var _ http.ResponseWriter = (*logResponseWriter)(nil)

// Логирование запросов
func (a *Application) loggMiddleware(next http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// обернём записывальщик
		logWriter := logResponseWriter{w, 0, 0}

		a.logger.Info(
			"incoming request",
			"method", r.Method,
			"path", r.URL.Path,
		)

		next.ServeHTTP(&logWriter, r)

		a.logger.Info(
			"outgoing response",
			"status_code", logWriter.satus,
			"size", logWriter.size,
			"duration", time.Since(start).String(),
		)
	}

	return http.HandlerFunc(fn)
}
