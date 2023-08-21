package batch

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eugene982/url-shortener/internal/handlers"
	"github.com/eugene982/url-shortener/internal/logger"
	"github.com/eugene982/url-shortener/internal/middleware"
	"github.com/eugene982/url-shortener/internal/model"
	"github.com/eugene982/url-shortener/internal/shortener"
)

// Генерирование короткой ссылки и сохранеине её во временном хранилище
// из запроса формата JSON
func NewBatchHandler(b handlers.BaseURLGetter, u handlers.Updater, s shortener.Shortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close() // Очищаем тело

		if ok, err := handlers.CheckContentType("application/json", r); !ok {
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

		// Получаем идентификатор пользователя из контекста
		userID, err := middleware.GetUserID(r)
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

			short, err := s.Short(batch.OriginalURL)
			if err != nil {
				logger.Warn("error get short url",
					"error", err)
				http.NotFound(w, r)
				return
			}

			response = append(response, model.BatchResponse{
				CorrelationID: batch.CorrelationID,
				ShortURL:      b.GetBaseURL() + short,
			})

			write = append(write, model.StoreData{
				ID:          batch.CorrelationID,
				UserID:      userID,
				ShortURL:    short,
				OriginalURL: batch.OriginalURL,
			})
		}

		if err = u.Update(r.Context(), write); err != nil {
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
}
