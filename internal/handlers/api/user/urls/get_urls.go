// Package urls - управление пользовательскими ссылоками
package urls

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eugene982/url-shortener/internal/handlers"
	"github.com/eugene982/url-shortener/internal/logger"
	"github.com/eugene982/url-shortener/internal/middleware"
	"github.com/eugene982/url-shortener/internal/model"
	"github.com/eugene982/url-shortener/proto"
)

// NewUserURLsHandler эндпоинт получения списка ссылок пользователя.
func NewUserURLsHandler(baseURL string, u handlers.UserURLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close() // Очищаем тело

		// Получаем идентификатор пользователя из контекста
		userID, err := middleware.GetUserID(r.Context())
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
				ShortURL:    baseURL + v.ShortURL,
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

// NewGRPCUserURLsHandler список ссылок пользователя
func NewGRPCUserURLsHandler(baseURL string, u handlers.UserURLGetter) handlers.GetUserURLsHandler {
	return func(ctx context.Context, in *proto.UserURLsRequest) (*proto.UserURLsResponse, error) {
		var response proto.UserURLsResponse

		// Получаем список ссылок пользователя
		urls, err := u.GetUserURLs(ctx, in.User)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		if len(urls) == 0 {
			response.Error = "no content"
			return &response, err
		}

		response.Response = make([]*proto.UserURLsResponse_UserURL, len(urls))
		for i, v := range urls {
			response.Response[i] = &proto.UserURLsResponse_UserURL{
				ShortUrl:    baseURL + v.ShortURL,
				OriginalUrl: v.OriginalURL,
			}
		}

		return &response, nil
	}
}
