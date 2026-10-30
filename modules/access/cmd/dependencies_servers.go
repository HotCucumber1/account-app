package main

import (
	"log/slog"

	grpcserver "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/distributed-programming-2026/lib/runtime/grpc"
	"github.com/distributed-programming-2026/lib/runtime/http"

	internalapi "access/api/server/internalapi"
	"access/internal/access/app"
	"access/internal/access/config"
	internalhandler "access/internal/access/infra/handlers/internalapi"
	publichandler "access/internal/access/infra/handlers/public"
)

func newHTTPServer(logger *slog.Logger, conf config.Env, svc *app.Service) (*http.Server, error) {
	register, err := publichandler.Register(svc, logger)
	if err != nil {
		return nil, err
	}

	return http.NewServer(http.Config{
		Address:      conf.HTTP.Address,
		GraceTimeout: conf.App.GraceTimeout,
	}, register)
}

func newGRPCServer(logger *slog.Logger, conf config.Env, svc *app.Service) (*grpc.Server, error) {
	return grpc.NewServer(grpc.Config{
		Address:      conf.GRPC.Address,
		GraceTimeout: conf.App.GraceTimeout,
	}, func(server *grpcserver.Server) error {
		internalapi.RegisterAccessServiceServer(server, internalhandler.NewHandler(svc, logger))
		reflection.Register(server)
		return nil
	})
}
