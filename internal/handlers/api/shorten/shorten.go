package shorten

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/eugene982/url-shortener/internal/handlers"
	"github.com/eugene982/url-shortener/internal/logger"
	"github.com/eugene982/url-shortener/internal/model"
	"github.com/eugene982/url-shortener/internal/shortener"
	"github.com/eugene982/url-shortener/internal/storage"
)

// Генерирование короткой ссылки и сохранеине её во временном хранилище
// из запроса формата JSON
func NewShortenHandler(b handlers.BaseURLGetter, setter handlers.Setter, sh shortener.Shortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close() // Очищаем тело

		if ok, err := handlers.CheckContentType("application/json", r); !ok {
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
		short, err := handlers.GetAndWriteShort(sh, setter, request.URL, r)

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

		response := model.ResponseShorten{
			Result: b.GetBaseURL() + short,
		}

		w.WriteHeader(http.StatusCreated)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error(fmt.Errorf("error encoding responce: %w", err))
			http.NotFound(w, r)
			return
		}
	}
}
