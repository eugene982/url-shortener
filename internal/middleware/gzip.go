package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/eugene982/url-shortener/internal/compress"
	"github.com/eugene982/url-shortener/internal/logger"
)

// Gzip прослойка упаковки и распаковка запросов gzip
func Gzip(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		// Клиент понимает gzip
		allowGZip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")

		// В запросе указан тип данных
		contentType := r.Header.Get("Content-Type")
		allowContentType := strings.Contains(contentType, "application/json") ||
			strings.Contains(contentType, "text/html")

		if allowGZip && allowContentType {
			cw, err := compress.NewGzipComressWriter(w)
			if err != nil {
				logger.Error(fmt.Errorf("failed to create gzip writer: %w", err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer cw.Close()
			w = cw
		}

		// Клиент отправил упакованные данные
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			cr, err := compress.NewGzipCompressReader(r.Body)
			if err != nil {
				logger.Error(fmt.Errorf("failed to create gzip reader: %w", err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer cr.Close()
			r.Body = cr
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
