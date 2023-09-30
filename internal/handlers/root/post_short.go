package root

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/eugene982/url-shortener/internal/handlers"
	"github.com/eugene982/url-shortener/internal/logger"
	"github.com/eugene982/url-shortener/internal/shortener"
	"github.com/eugene982/url-shortener/internal/storage"
	"github.com/eugene982/url-shortener/proto"
)

// NewCreateShortHandler эндпоинт получения короткой ссылки.
// Генерирование короткой ссылки и сохранеине её в хранилище.
func NewCreateShortHandler(baseURL string, setter handlers.Setter, sh shortener.Shortener) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		body, err := io.ReadAll(r.Body)
		defer r.Body.Close() // Вроде как надо закрывать если что-то там есть...
		if err != nil {
			logger.Error(fmt.Errorf("error read body: %w", err))
			http.NotFound(w, r)
			return
		}

		addr := string(body)
		short, err := handlers.GetAndWriteShort(sh, setter, addr, r)
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

		// linter
		if _, err := io.WriteString(w, baseURL+short); err != nil {
			logger.Error(err)
			http.NotFound(w, r)
		}
	}
}

// NewGRPCCreateShortHandler эндпоинт получения короткой ссылки grpc.
func NewGRPCCreateShortHandler(baseURL string, setter handlers.Setter, sh shortener.Shortener) handlers.CreateShortHandler {

	return func(ctx context.Context, in *proto.CreateShortRequest) (*proto.CreateShortResponse, error) {
		var (
			response proto.CreateShortResponse
			err      error
		)
		response.ShortUrl, err = handlers.GetAndWriteUserShort(ctx, sh, setter, in.OriginalUrl, in.User)
		if err != nil {
			return nil, err
		}
		return &response, nil
	}
}
