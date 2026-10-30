package main

import (
	"log/slog"

	grpcserver "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/distributed-programming-2026/lib/common/grpcx"
	gatewayapi "github.com/distributed-programming-2026/lib/gateway/interceptor"
	"github.com/distributed-programming-2026/lib/runtime/grpc"
	"github.com/distributed-programming-2026/lib/runtime/http"

	internalapi "auth/api/server/internalapi"
	"auth/internal/auth/config"
	gatewayhandler "auth/internal/auth/infra/handlers/gateway"
	internalhandler "auth/internal/auth/infra/handlers/internalapi"
	publichandler "auth/internal/auth/infra/handlers/public"
)

func newHTTPServer(logger *slog.Logger, conf config.Env, s *services) (*http.Server, error) {
	register, err := publichandler.Register(publichandler.NewHandler(s.service), logger)
	if err != nil {
		return nil, err
	}

	return http.NewServer(http.Config{
		Address:      conf.HTTP.Address,
		GraceTimeout: conf.App.GraceTimeout,
	}, register)
}

func newGRPCServer(logger *slog.Logger, conf config.Env, s *services) (*grpc.Server, error) {
	return grpc.NewServer(grpc.Config{
		Address:      conf.GRPC.Address,
		GraceTimeout: conf.App.GraceTimeout,
		Options: []grpcserver.ServerOption{
			grpcserver.UnaryInterceptor(grpcx.ErrorInterceptor(logger, internalhandler.ErrorCodes()...)),
		},
	}, func(server *grpcserver.Server) error {
		internalapi.RegisterAuthServiceServer(server, internalhandler.NewHandler(s.service))
		gatewayapi.RegisterInterceptorServer(server, gatewayhandler.NewHandler(s.service))
		reflection.Register(server)
		return nil
	})
}
