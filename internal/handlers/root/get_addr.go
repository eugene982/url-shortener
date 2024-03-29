// Package root ручки полученя адреса по короткой ссылке
package root

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/eugene982/url-shortener/gen/go/proto/v1"
	"github.com/eugene982/url-shortener/internal/handlers"
	"github.com/eugene982/url-shortener/internal/logger"
	"github.com/eugene982/url-shortener/internal/storage"
)

// NewFindAddrHandler эндпоинт получение полного адреса по короткой ссылке.
func NewFindAddrHandler(g handlers.AddrGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		short := chi.URLParam(r, "short")
		data, err := g.GetAddr(r.Context(), short)
		if err != nil {
			if errors.Is(storage.ErrAddressNotFound, err) {
				logger.Info(err.Error(), "short", short)
			} else {
				logger.Error(err, "short", short)
			}
			http.NotFound(w, r)
			return
		}

		if data.DeletedFlag {
			http.Error(w, "410 Gone", http.StatusGone)
			return
		}

		w.Header().Set("Location", data.OriginalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

// NewGRPCFindAddrHandler получение полного адреса по короткой ссылке для gRPC
func NewGRPCFindAddrHandler(g handlers.AddrGetter) handlers.FindAddrHandler {
	return func(ctx context.Context, in *proto.FindAddrRequest) (*proto.FindAddrResponse, error) {
		var responce proto.FindAddrResponse

		data, err := g.GetAddr(ctx, in.ShortUrl)
		if err == nil {
			if data.DeletedFlag {
				return nil, status.Error(codes.NotFound, "Delete")
			} else {
				responce.OriginalUrl = data.OriginalURL
			}
		} else if errors.Is(storage.ErrAddressNotFound, err) {
			logger.Info(err.Error(), "short", in.ShortUrl)
			return nil, status.Error(codes.NotFound, err.Error())
		} else {
			logger.Error(err, "short", in.ShortUrl)
			return nil, err
		}
		return &responce, nil
	}
}
