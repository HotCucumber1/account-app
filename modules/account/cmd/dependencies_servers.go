package main

import (
	"log/slog"

	grpcserver "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/distributed-programming-2026/lib/common/grpcx"
	"github.com/distributed-programming-2026/lib/runtime/grpc"
	"github.com/distributed-programming-2026/lib/runtime/http"

	internalapi "account/api/server/internalapi"
	"account/internal/account/config"
	internalhandler "account/internal/account/infra/handlers/internalapi"
	publichandler "account/internal/account/infra/handlers/public"
)

func newHTTPServer(logger *slog.Logger, conf config.Env, s *services) (*http.Server, error) {
	register, err := publichandler.Register(
		publichandler.NewHandler(*s.service, s.queryService),
		logger,
	)
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
		internalapi.RegisterAccountServiceServer(server, internalhandler.NewHandler(s.service, s.queryService))
		reflection.Register(server)
		return nil
	})
}
