package urls

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eugene982/url-shortener/internal/handlers"
	"github.com/eugene982/url-shortener/internal/logger"
	"github.com/eugene982/url-shortener/internal/middleware"
	"github.com/eugene982/url-shortener/internal/model"
)

// Список ссылок пользователя
func NewUserURLsHandler(b handlers.BaseURLGetter, u handlers.UserURLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close() // Очищаем тело

		// Получаем идентификатор пользователя из контекста
		userID, err := middleware.GetUserID(r)
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Получаем список ссылок пользователя
		urls, err := u.GetUserURLs(r.Context(), userID)
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(urls) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		responce := make([]model.UserURLResponse, len(urls))
		for i, v := range urls {
			responce[i] = model.UserURLResponse{
				ShortURL:    b.GetBaseURL() + v.ShortURL,
				OriginalURL: v.OriginalURL,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(responce); err != nil {
			logger.Error(fmt.Errorf("error encoding responce: %w", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
