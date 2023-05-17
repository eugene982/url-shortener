// Роутер - на основе внешней библиотеки "chi"
package app

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// Возвращает роутер
func (a *Application) NewRouter() http.Handler {

	r := chi.NewRouter()

	r.Get("/{short}", a.findAddr)
	r.Post("/", a.createShort)

	// во всех остальных случаях 404
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		writeErrNotFound(fmt.Errorf("method not allowed"), w, r)
	})

	return r
}

// Получение полного адреса по короткой ссылке
func (a *Application) findAddr(w http.ResponseWriter, r *http.Request) {

	short := chi.URLParam(r, "short")
	addr, ok := a.store.GetAddr(short)
	if !ok {
		writeErrNotFound(fmt.Errorf("short address %s not found", short), w, r)
		return
	}

	w.Header().Set("Location", addr)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// Генерирование короткой ссылки и сохранеине её во временном хранилище
func (a *Application) createShort(w http.ResponseWriter, r *http.Request) {

	err := checkContentType("text/plain", r)
	if err != nil {
		writeErrNotFound(err, w, r)
		return
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close() // Вроде как надо закрывать если что-то там есть...
	if err != nil {
		writeErrNotFound(err, w, r)
		return
	}

	short, err := a.getAndWriteShort(string(body), w, r)
	if err != nil {
		writeErrNotFound(err, w, r)
		return
	}

	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, short)
}

// при ошибке всегда возвращаем 404
func writeErrNotFound(err error, w http.ResponseWriter, r *http.Request) {
	log.Println("error:", err)
	http.NotFound(w, r)
}

// проверка заголовка на формат
func checkContentType(value string, r *http.Request) error {
	if strings.Contains(r.Header.Get("Content-Type"), value) {
		return nil
	}
	return fmt.Errorf("Content-Type: %s not found", value)
}

// ищем или пытаемся создать короткую ссылку
func (a *Application) getAndWriteShort(addr string, w http.ResponseWriter, r *http.Request) (string, error) {

	if addr == "" {
		return "", fmt.Errorf("address is empty")
	}

	short := a.shortener.Short(addr)
	if short == "" {
		return "", fmt.Errorf("short url generation error")
	}

	a.store.Set(addr, short)
	return a.baseURL + short, nil
}
