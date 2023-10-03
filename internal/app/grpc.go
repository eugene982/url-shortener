package app

import (
	"context"
	"net"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"github.com/eugene982/url-shortener/gen/go/proto"
	"github.com/eugene982/url-shortener/internal/handlers"
	"github.com/eugene982/url-shortener/internal/handlers/api/shorten/batch"
	"github.com/eugene982/url-shortener/internal/handlers/api/user/urls"
	"github.com/eugene982/url-shortener/internal/handlers/ping"
	"github.com/eugene982/url-shortener/internal/handlers/root"
)

type protoServer struct {
	proto.UnimplementedShortenerServer

	pingHandler        handlers.PingHandler
	findHandler        handlers.FindAddrHandler
	createHandler      handlers.CreateShortHandler
	batchHandler       handlers.BatchShortHandler
	userURLsHandler    handlers.GetUserURLsHandler
	delUserURLsHandler handlers.DelUserURLsHandler
}

type GRPCServer struct {
	listen net.Listener
	server *grpc.Server
	proto  *protoServer
}

func NewGRPCServer(a *Application, addr string) (*GRPCServer, error) {
	var (
		srv GRPCServer
		err error
	)

	// определяем адрес сервера
	srv.listen, err = net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	// создаём gRPC-сервер без зарегистрированной службы
	srv.server = grpc.NewServer()

	srv.proto = &protoServer{
		pingHandler:        ping.NewGRPCPingHandler(a.store),
		findHandler:        root.NewGRPCFindAddrHandler(a.store),
		createHandler:      root.NewGRPCCreateShortHandler(a.baseURL, a.store, a.shortener),
		batchHandler:       batch.NewGRPCBatchHandler(a.baseURL, a.store, a.shortener),
		userURLsHandler:    urls.NewGRPCUserURLsHandler(a.baseURL, a.store),
		delUserURLsHandler: urls.NewGRPCDeleteURLsHandlers(a),
	}

	// регистрируем сервис
	proto.RegisterShortenerServer(srv.server, srv.proto)

	return &srv, nil
}

func (s *GRPCServer) Start() error {
	return s.server.Serve(s.listen)
}

func (s *protoServer) Ping(ctx context.Context, in *empty.Empty) (*proto.PingResponse, error) {
	return s.pingHandler(ctx, in)
}

func (s *protoServer) FindAddr(ctx context.Context, in *proto.FindAddrRequest) (*proto.FindAddrResponse, error) {
	return s.findHandler(ctx, in)
}

func (s *protoServer) CreateShort(ctx context.Context, in *proto.CreateShortRequest) (*proto.CreateShortResponse, error) {
	return s.createHandler(ctx, in)
}

func (s *protoServer) BatchShort(ctx context.Context, in *proto.BatchRequest) (*proto.BatchResponse, error) {
	return s.batchHandler(ctx, in)
}

func (s *protoServer) GetUserURLs(ctx context.Context, in *proto.UserURLsRequest) (*proto.UserURLsResponse, error) {
	return s.userURLsHandler(ctx, in)
}

func (s *protoServer) DelUserURLs(ctx context.Context, in *proto.DelUserURLsRequest) (*empty.Empty, error) {
	return s.delUserURLsHandler(ctx, in)
}
