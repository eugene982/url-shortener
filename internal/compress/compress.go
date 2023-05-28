package compress

import (
	"compress/gzip"
	"io"
	"net/http"
)

type GzipComressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// Проверка на совместимость с интерфейсом http.ResponseWriter
var _ http.ResponseWriter = (*GzipComressWriter)(nil)

// Конструктор
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

// Досылка всех данных из буфера
func (cw *GzipComressWriter) Close() error {
	return cw.zw.Close()
}

// Структура для чтения упакованных данных
type GzipCompressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// Проверка на совместимость с интерфейсом io.ReadCloser
var _ io.ReadCloser = (*GzipCompressReader)(nil)

// Конструктор
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

// Read implements http.ReadCloser
func (cr *GzipCompressReader) Close() error {
	if err := cr.r.Close(); err != nil {
		return err
	}
	return cr.zr.Close()
}
