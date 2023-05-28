// Роутер - на основе внешней библиотеки "chi"
package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/eugene982/url-shortener/internal/model"
)

// Возвращает роутер
func (a *Application) NewRouter() http.Handler {

	r := chi.NewRouter()

	r.Use(a.loggMiddleware) // прослойка логирования
	r.Use(a.gzipMiddleware) // прослойка сжатия

	r.Get("/{short}", a.findAddr)
	r.Post("/", a.createShort)
	r.Post("/api/shorten", a.createAPIShorten)

	// во всех остальных случаях 404
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		a.logger.Warn("not allowed",
			"method", r.Method)
		http.NotFound(w, r)
	})

	return r
}

// Получение полного адреса по короткой ссылке
func (a *Application) findAddr(w http.ResponseWriter, r *http.Request) {

	short := chi.URLParam(r, "short")
	addr, ok := a.store.GetAddr(short)
	if !ok {
		a.logger.Warn("not found",
			"short", short)
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Location", addr)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// Генерирование короткой ссылки и сохранеине её во временном хранилище
func (a *Application) createShort(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close() // Вроде как надо закрывать если что-то там есть...
	if err != nil {
		a.logger.Error(fmt.Errorf("error read body: %w", err))
		http.NotFound(w, r)
		return
	}

	short, err := a.getAndWriteShort(string(body), w, r)
	if err != nil {
		a.logger.Warn(err.Error())
		http.NotFound(w, r)
		return
	}

	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, short)
}

// Генерирование короткой ссылки и сохранеине её во временном хранилище
// из запроса формата JSON
func (a *Application) createAPIShorten(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // Очищаем тело

	if ok, err := checkContentType("application/json", r); !ok {
		a.logger.Warn(err.Error())
		http.NotFound(w, r)
		return
	}

	// получаем тело ответа и проверяем его
	var request model.RequestShorten
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		a.logger.Warn("wrong body",
			"error", err)
		http.NotFound(w, r)
		return
	}

	if ok, err := request.IsValid(); !ok {
		a.logger.Warn("request is not valid",
			"error", err)
		http.NotFound(w, r)
		return
	}

	//	подготовка ответа
	var response model.ResponseShorten
	response.Result, err = a.getAndWriteShort(request.URL, w, r)
	if err != nil {
		a.logger.Warn(err.Error())
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		a.logger.Error(fmt.Errorf("error encoding responce: %w", err))
		http.NotFound(w, r)
		return
	}
}

// проверка заголовка на формат
func checkContentType(value string, r *http.Request) (bool, error) {
	if strings.Contains(r.Header.Get("Content-Type"), value) {
		return true, nil
	}
	return false, fmt.Errorf("Content-Type: %s not found", value)
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

	// запись в файловое хранилище
	if err := a.fileStorage.Append(addr, short); err != nil {
		return "", err
	}

	a.store.Set(addr, short)
	return a.baseURL + short, nil
}
