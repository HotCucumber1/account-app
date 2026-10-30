package internalapi

import (
	"context"
	"log/slog"

	"github.com/distributed-programming-2026/go-sdk/pkg/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	billinginternalapi "billing/api/server/internalapi"
	"billing/internal/billing/app"
)

func NewHandler(svc *app.Service, logger *slog.Logger) billinginternalapi.BillingServiceServer {
	return &handler{
		svc:    svc,
		logger: logger,
	}
}

type handler struct {
	svc    *app.Service
	logger *slog.Logger
	billinginternalapi.UnimplementedBillingServiceServer
}

func (h handler) Echo(ctx context.Context, req *billinginternalapi.EchoRequest) (*billinginternalapi.EchoResponse, error) {
	id, err := h.svc.Echo(ctx, req.Body)
	if err != nil {
		h.logger.ErrorContext(ctx, "echo failed", logging.Error(err))
		return nil, status.Error(codes.Internal, "echo failed")
	}

	return &billinginternalapi.EchoResponse{
		OperationID: id.String(),
	}, nil
}
