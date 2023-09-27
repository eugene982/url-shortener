package middleware

import (
	"net"
	"net/http"

	"github.com/eugene982/url-shortener/internal/logger"
)

type TrustedSubnet string

func (t TrustedSubnet) Serve(next http.Handler) http.Handler {

	_, subnet, err := net.ParseCIDR(string(t))
	if err != nil {
		logger.Error(err)
	}

	fn := func(w http.ResponseWriter, r *http.Request) {
		realIP := r.Header.Get("X-Real-IP")

		if t == "" || realIP == "" {
			http.Error(w, "403 Forbidden", http.StatusForbidden)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		real := net.ParseIP(realIP)
		if real == nil {
			http.Error(w, "403 Forbidden", http.StatusForbidden)
			return
		}

		if !subnet.Contains(real) {
			http.Error(w, "403 Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
