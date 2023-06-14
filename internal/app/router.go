// Роутер - на основе внешней библиотеки "chi"
package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/eugene982/url-shortener/internal/logger"
	"github.com/eugene982/url-shortener/internal/middleware"
	"github.com/eugene982/url-shortener/internal/model"
	"github.com/eugene982/url-shortener/internal/storage"
)

// Возвращает роутер
func (a *Application) NewRouter() http.Handler {

	r := chi.NewRouter()

	r.Use(middleware.Log)  // прослойка логирования
	r.Use(middleware.Gzip) // прослойка сжатия

	r.Get("/ping", a.handlerPing)
	r.Get("/{short}", a.handlerFindAddr)

	r.Post("/", a.handlerCreateShort)
	r.Post("/api/shorten", a.handlerAPIShorten)
	r.Post("/api/shorten/batch", a.handlerAPIBatch)

	// во всех остальных случаях 404
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		logger.Warn("not allowed",
			"method", r.Method)
		http.NotFound(w, r)
	})

	return r
}

// Получение полного адреса по короткой ссылке
func (a *Application) handlerFindAddr(w http.ResponseWriter, r *http.Request) {

	short := chi.URLParam(r, "short")
	addr, err := a.store.GetAddr(r.Context(), short)
	if err != nil {
		if errors.Is(storage.ErrAddressNotFound, err) {
			logger.Info(err.Error(), "short", short)
		} else {
			logger.Error(err, "short", short)
		}
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Location", addr)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// Генерирование короткой ссылки и сохранеине её во временном хранилище
func (a *Application) handlerCreateShort(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close() // Вроде как надо закрывать если что-то там есть...
	if err != nil {
		logger.Error(fmt.Errorf("error read body: %w", err))
		http.NotFound(w, r)
		return
	}

	addr := string(body)
	short, err := a.getAndWriteShort(addr, w, r)
	if err == nil {
		w.WriteHeader(http.StatusCreated)

	} else if errors.Is(err, storage.ErrAddressConflict) {
		logger.Warn(err.Error(),
			"url", addr)
		w.WriteHeader(http.StatusConflict)

	} else {
		logger.Error(err)
		http.NotFound(w, r)
		return
	}

	io.WriteString(w, short)
}

// Генерирование короткой ссылки и сохранеине её во временном хранилище
// из запроса формата JSON
func (a *Application) handlerAPIShorten(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // Очищаем тело

	if ok, err := checkContentType("application/json", r); !ok {
		logger.Warn(err.Error())
		http.NotFound(w, r)
		return
	}

	// получаем тело ответа и проверяем его
	var request model.RequestShorten
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		logger.Warn("wrong body",
			"error", err)
		http.NotFound(w, r)
		return
	}

	if ok, err := request.IsValid(); !ok {
		logger.Warn("request is not valid",
			"error", err)
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	//	подготовка ответа
	var response model.ResponseShorten
	response.Result, err = a.getAndWriteShort(request.URL, w, r)

	if err == nil {
		w.WriteHeader(http.StatusCreated)

	} else if errors.Is(err, storage.ErrAddressConflict) {
		logger.Warn(err.Error(),
			"url", request.URL)
		w.WriteHeader(http.StatusConflict)

	} else {
		logger.Warn("error write short url",
			"url", request.URL,
			"err", err)
		http.NotFound(w, r)
		return
	}

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error(fmt.Errorf("error encoding responce: %w", err))
		http.NotFound(w, r)
		return
	}
}

// Генерирование короткой ссылки и сохранеине её во временном хранилище
// из запроса формата JSON
func (a *Application) handlerAPIBatch(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // Очищаем тело

	if ok, err := checkContentType("application/json", r); !ok {
		logger.Warn(err.Error())
		http.NotFound(w, r)
		return
	}

	// получаем тело ответа и проверяем его
	request := make([]model.BatchRequest, 0) // сюда прочитаем запрос
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		logger.Warn("wrong body",
			"error", err)
		http.NotFound(w, r)
		return
	}

	response := make([]model.BatchResponse, 0, len(request)) // подготовка ответа
	write := make([]model.StoreData, 0, len(request))        // это положим в хранилище

	for _, batch := range request {

		if ok, err := batch.IsValid(); !ok {
			logger.Warn("request is not valid",
				"error", err)
			http.NotFound(w, r)
			return
		}

		short, err := a.shortener.Short(batch.OriginalURL)
		if err != nil {
			logger.Warn("error get short url",
				"error", err)
			http.NotFound(w, r)
			return
		}

		response = append(response, model.BatchResponse{
			CorrelationID: batch.CorrelationID,
			ShortURL:      a.baseURL + short,
		})

		write = append(write, model.StoreData{
			ID:          batch.CorrelationID,
			ShortURL:    short,
			OriginalURL: batch.OriginalURL,
		})
	}

	if err = a.store.Update(r.Context(), write); err != nil {
		logger.Error(fmt.Errorf("error write data in storage: %w", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error(fmt.Errorf("error encoding responce: %w", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Проверка соединения
func (a *Application) handlerPing(w http.ResponseWriter, r *http.Request) {
	err := a.store.Ping(r.Context())
	if err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "pong")
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

	short, err := a.shortener.Short(addr)
	if err != nil {
		return "", err
	}

	// запись в файловое хранилище
	err = a.store.Set(r.Context(), addr, short)
	return a.baseURL + short, err
}
