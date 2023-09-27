package stats

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eugene982/url-shortener/internal/handlers"
	"github.com/eugene982/url-shortener/internal/logger"
	"github.com/eugene982/url-shortener/internal/model"
)

func NewStatsHandler(s handlers.StatsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			resp model.StatsResponse
			err  error
		)

		resp.URLs, resp.Users, err = s.Stats(r.Context())
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Error(fmt.Errorf("error encoding responce: %w", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
