package app

import (
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"

	"github.com/eugene982/url-shortener/internal/logger"
	"github.com/eugene982/url-shortener/internal/middleware"

	"github.com/eugene982/url-shortener/internal/handlers/api/intrnl/stats"
	"github.com/eugene982/url-shortener/internal/handlers/api/shorten"
	"github.com/eugene982/url-shortener/internal/handlers/api/shorten/batch"
	"github.com/eugene982/url-shortener/internal/handlers/api/user/urls"
	"github.com/eugene982/url-shortener/internal/handlers/ping"
	"github.com/eugene982/url-shortener/internal/handlers/root"
)

// NewRouter функция создаёт и возвращает роутер.
func NewRouter(a *Application) http.Handler {

	r := chi.NewRouter()

	r.Use(middleware.Log)  // прослойка логирования
	r.Use(middleware.Gzip) // прослойка сжатия

	// Прослойка авторизации
	r.Use(middleware.Auth)

	r.Get("/ping", ping.NewPingHandler(a.store))
	r.Get("/{short}", root.NewFindAddrHandler(a.store))

	r.Post("/", root.NewCreateShortHandler(a.baseURL, a.store, a.shortener))
	r.Post("/api/shorten", shorten.NewShortenHandler(a.baseURL, a.store, a.shortener))
	r.Post("/api/shorten/batch", batch.NewBatchHandler(a.baseURL, a.store, a.shortener))

	r.Get("/api/user/urls", urls.NewUserURLsHandler(a.baseURL, a.store))
	r.Delete("/api/user/urls", urls.NewDeleteURLsHandlers(a))

	r.Group(func(r chi.Router) {
		r.Use(middleware.TrustedSubnet(a.trustedSubnet).Serve)
		r.Get("/api/internal/stats", stats.NewStatsHandler(a.store))
	})

	// во всех остальных случаях 404
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		logger.Warn("not allowed",
			"method", r.Method)
		http.NotFound(w, r)
	})

	return r
}

// NewProfRouter создаёт маршрутизатор для профилирования
func newProfRouter() http.Handler {

	r := chi.NewRouter()
	r.Use(middleware.Log) // прослойка логирования

	r.HandleFunc("/debug/pprof/*", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return r

}
