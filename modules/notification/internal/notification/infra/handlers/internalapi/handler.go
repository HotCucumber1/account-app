package internalapi

import (
	"context"
	"log/slog"

	"github.com/distributed-programming-2026/go-sdk/pkg/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	notificationinternalapi "notification/api/server/internalapi"
	"notification/internal/notification/app"
)

func NewHandler(svc *app.Service, logger *slog.Logger) notificationinternalapi.NotificationServiceServer {
	return &handler{
		svc:    svc,
		logger: logger,
	}
}

type handler struct {
	svc    *app.Service
	logger *slog.Logger
	notificationinternalapi.UnimplementedNotificationServiceServer
}

func (h handler) Echo(ctx context.Context, req *notificationinternalapi.EchoRequest) (*notificationinternalapi.EchoResponse, error) {
	id, err := h.svc.Echo(ctx, req.Body)
	if err != nil {
		h.logger.ErrorContext(ctx, "echo failed", logging.Error(err))
		return nil, status.Error(codes.Internal, "echo failed")
	}

	return &notificationinternalapi.EchoResponse{
		OperationID: id.String(),
	}, nil
}
