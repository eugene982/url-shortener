package compress

import (
	"compress/gzip"
	"io"
	"net/http"
)

// GzipComressWriter запись сжатых данных в ответ
type GzipComressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// Проверка на совместимость с интерфейсом http.ResponseWriter
var _ http.ResponseWriter = (*GzipComressWriter)(nil)

// NewGzipComressWriter функция-конструктор
func NewGzipComressWriter(w http.ResponseWriter) (*GzipComressWriter, error) {
	zw, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}
	return &GzipComressWriter{w, zw}, nil
}

// Header implements http.ResponseWriter
func (cw *GzipComressWriter) Header() http.Header {
	return cw.w.Header()
}

// Write implements http.ResponseWriter
func (cw *GzipComressWriter) Write(p []byte) (n int, err error) {
	return cw.zw.Write(p)
}

// WriteHeader implements http.ResponseWriter
func (cw *GzipComressWriter) WriteHeader(statusCode int) {
	cw.w.Header().Set("Content-Encoding", "gzip")
	cw.w.WriteHeader(statusCode)
}

// Close - досылка всех данных из буфера, закрытие.
func (cw *GzipComressWriter) Close() error {
	return cw.zw.Close()
}

// GzipCompressReader Структура для чтения упакованных данных.
type GzipCompressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// Проверка на совместимость с интерфейсом io.ReadCloser
var _ io.ReadCloser = (*GzipCompressReader)(nil)

// NewGzipCompressReader функция-конструктор
func NewGzipCompressReader(r io.ReadCloser) (*GzipCompressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &GzipCompressReader{r, zr}, nil
}

// Read implements http.ReadCloser
func (cr *GzipCompressReader) Read(p []byte) (n int, err error) {
	return cr.zr.Read(p)
}

// Close implements http.ReadCloser
func (cr *GzipCompressReader) Close() error {
	if err := cr.r.Close(); err != nil {
		return err
	}
	return cr.zr.Close()
}
