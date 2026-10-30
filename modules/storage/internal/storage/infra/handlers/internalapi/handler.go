package internalapi

import (
	"context"
	"log/slog"

	"github.com/distributed-programming-2026/go-sdk/pkg/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	storageinternalapi "storage/api/server/internalapi"
	"storage/internal/storage/app"
)

func NewHandler(svc *app.Service, logger *slog.Logger) storageinternalapi.StorageServiceServer {
	return &handler{
		svc:    svc,
		logger: logger,
	}
}

type handler struct {
	svc    *app.Service
	logger *slog.Logger
	storageinternalapi.UnimplementedStorageServiceServer
}

func (h handler) Echo(ctx context.Context, req *storageinternalapi.EchoRequest) (*storageinternalapi.EchoResponse, error) {
	id, err := h.svc.Echo(ctx, req.Body)
	if err != nil {
		h.logger.ErrorContext(ctx, "echo failed", logging.Error(err))
		return nil, status.Error(codes.Internal, "echo failed")
	}

	return &storageinternalapi.EchoResponse{
		OperationID: id.String(),
	}, nil
}
