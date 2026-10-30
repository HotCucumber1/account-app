package main

import (
	"log/slog"

	grpcserver "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/distributed-programming-2026/lib/common/grpcx"
	"github.com/distributed-programming-2026/lib/runtime/grpc"
	"github.com/distributed-programming-2026/lib/runtime/http"

	internalapi "session/api/server/internalapi"
	"session/internal/session/config"
	internalhandler "session/internal/session/infra/handlers/internalapi"
)

func newHTTPServer(_ *slog.Logger, conf config.Env, _ *services) (*http.Server, error) {
	return http.NewServer(http.Config{
		Address:      conf.HTTP.Address,
		GraceTimeout: conf.App.GraceTimeout,
	})
}

func newGRPCServer(logger *slog.Logger, conf config.Env, s *services) (*grpc.Server, error) {
	return grpc.NewServer(grpc.Config{
		Address:      conf.GRPC.Address,
		GraceTimeout: conf.App.GraceTimeout,
		Options: []grpcserver.ServerOption{
			grpcserver.UnaryInterceptor(grpcx.ErrorInterceptor(logger, internalhandler.ErrorCodes()...)),
		},
	}, func(server *grpcserver.Server) error {
		internalapi.RegisterSessionServiceServer(server, internalhandler.NewHandler(s.service, s.queryService))
		reflection.Register(server)
		return nil
	})
}
