package urls

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/eugene982/url-shortener/gen/go/proto/v1"
	"github.com/eugene982/url-shortener/internal/handlers"
	"github.com/eugene982/url-shortener/internal/logger"
	"github.com/eugene982/url-shortener/internal/middleware"
)

// NewDeleteURLsHandlers эндпоинт удаление ссылок пользователя.
// Асинхронный.
func NewDeleteURLsHandlers(d handlers.UserShortAsyncDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close() // Очищаем тело

		if ok, err := handlers.CheckContentType("application/json", r); !ok {
			logger.Warn(err.Error())
			http.NotFound(w, r)
			return
		}

		// получаем тело ответа и проверяем его
		request := make([]string, 0) // сюда прочитаем запрос
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			logger.Warn("wrong body",
				"error", err)
			http.NotFound(w, r)
			return
		}

		// Получаем идентификатор пользователя из контекста
		userID, err := middleware.GetUserID(r.Context())
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		d.DeleteUserShortAsync(userID, request)

		w.WriteHeader(http.StatusAccepted)
	}
}

// NewGRPCDeleteURLsHandlers асинхронное удаление ссылок пользователя
func NewGRPCDeleteURLsHandlers(d handlers.UserShortAsyncDeleter) handlers.DelUserURLsHandler {

	return func(ctx context.Context, in *proto.DelUserURLsRequest) (*empty.Empty, error) {
		d.DeleteUserShortAsync(in.User, in.ShortUrl)
		return &empty.Empty{}, nil
	}
}
