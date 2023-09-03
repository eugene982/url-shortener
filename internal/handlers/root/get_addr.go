package root

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

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
