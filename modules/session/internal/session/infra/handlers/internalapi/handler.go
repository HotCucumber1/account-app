package internalapi

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/distributed-programming-2026/lib/common/grpcx"

	sessioninternalapi "session/api/server/internalapi"
	"session/internal/session/app"
	appquery "session/internal/session/app/query"
	appservice "session/internal/session/app/service"
	"session/internal/session/domain"
)

func ErrorCodes() []grpcx.ErrorCode {
	return []grpcx.ErrorCode{
		{Err: domain.ErrSessionNotFound, Code: codes.NotFound},
		{Err: domain.ErrInvalidRevokeReason, Code: codes.InvalidArgument},
		{Err: domain.ErrInvalidSessionParams, Code: codes.InvalidArgument},
	}
}

func NewHandler(service *appservice.SessionService, queryService appquery.SessionQueryService) sessioninternalapi.SessionServiceServer {
	return &handler{
		service:      service,
		queryService: queryService,
	}
}

type handler struct {
	service      *appservice.SessionService
	queryService appquery.SessionQueryService
	sessioninternalapi.UnimplementedSessionServiceServer
}

func (h handler) CreateSession(ctx context.Context, req *sessioninternalapi.CreateSessionRequest) (*sessioninternalapi.CreateSessionResponse, error) {
	accountID, err := grpcx.ParseUUID("AccountID", req.AccountID)
	if err != nil {
		return nil, err
	}
	created, err := h.service.CreateSession(ctx, app.CreateSessionInput{
		AccountID: accountID,
		UserAgent: req.UserAgent,
		IPAddress: req.IPAddress,
	})
	if err != nil {
		return nil, err
	}
	return &sessioninternalapi.CreateSessionResponse{
		SessionID: created.SessionID.String(),
		ExpiresAt: timestamppb.New(created.ExpiresAt),
		NewDevice: created.NewDevice,
	}, nil
}

func (h handler) GetSession(ctx context.Context, req *sessioninternalapi.GetSessionRequest) (*sessioninternalapi.Session, error) {
	id, err := grpcx.ParseUUID("SessionID", req.SessionID)
	if err != nil {
		return nil, err
	}
	session, err := h.queryService.GetSession(ctx, id)
	if err != nil {
		return nil, err
	}
	return &sessioninternalapi.Session{
		ID:           session.ID.String(),
		AccountID:    session.AccountID.String(),
		DeviceID:     session.DeviceID.String(),
		UserAgent:    session.UserAgent,
		IPAddress:    session.IPAddress,
		Status:       session.Status,
		CreatedAt:    timestamppb.New(session.CreatedAt),
		ExpiresAt:    timestamppb.New(session.ExpiresAt),
		RevokeReason: session.RevokeReason,
	}, nil
}

func (h handler) ValidateSession(ctx context.Context, req *sessioninternalapi.ValidateSessionRequest) (*sessioninternalapi.ValidateSessionResponse, error) {
	id, err := grpcx.ParseUUID("SessionID", req.SessionID)
	if err != nil {
		return nil, err
	}
	validation, err := h.queryService.ValidateSession(ctx, id)
	if err != nil || !validation.Valid {
		return &sessioninternalapi.ValidateSessionResponse{}, err
	}
	return &sessioninternalapi.ValidateSessionResponse{
		Valid:     true,
		AccountID: validation.AccountID.String(),
		ExpiresAt: timestamppb.New(validation.ExpiresAt),
	}, nil
}

func (h handler) KillSession(ctx context.Context, req *sessioninternalapi.KillSessionRequest) (*sessioninternalapi.KillSessionResponse, error) {
	id, err := grpcx.ParseUUID("SessionID", req.SessionID)
	if err != nil {
		return nil, err
	}
	return &sessioninternalapi.KillSessionResponse{}, h.service.RevokeSession(ctx, app.KillSessionInput{SessionID: id, Reason: req.Reason})
}

func (h handler) KillAccountSessions(ctx context.Context, req *sessioninternalapi.KillAccountSessionsRequest) (*sessioninternalapi.KillAccountSessionsResponse, error) {
	accountID, err := grpcx.ParseUUID("AccountID", req.AccountID)
	if err != nil {
		return nil, err
	}
	exceptSessionID, err := grpcx.ParseUUIDOrNil("ExceptSessionID", req.ExceptSessionID)
	if err != nil {
		return nil, err
	}
	killed, err := h.service.KillAccountSessions(ctx, app.KillAccountSessionsInput{
		AccountID:       accountID,
		ExceptSessionID: exceptSessionID,
		Reason:          req.Reason,
	})
	if err != nil {
		return nil, err
	}
	return &sessioninternalapi.KillAccountSessionsResponse{Killed: int64(killed)}, nil
}
