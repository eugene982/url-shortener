package ping

import (
	"context"
	"fmt"
	"net/http"

	"github.com/eugene982/url-shortener/internal/handlers"
	"github.com/eugene982/url-shortener/internal/logger"
	"github.com/eugene982/url-shortener/proto"
)

// NewPingHandler эндпоинт проверки соединения.
func NewPingHandler(p handlers.Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := p.Ping(r.Context())
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "pong")
	}
}

// NewGRPCPingHandler проверка соединения по протоколу grpc
func NewGRPCPingHandler(p handlers.Pinger) handlers.PingHandler {
	return func(ctx context.Context, in *proto.PingRequest) (*proto.PingResponse, error) {
		err := p.Ping(ctx)
		if err != nil {
			logger.Error(err)
			return nil, err
		}
		return &proto.PingResponse{Message: "pong"}, nil
	}
}
