package session

import (
	"context"
	"time"
	"uuid"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/distributed-programming-2026/lib/common/grpcx"

	"auth/internal/auth/app"
	sessioninternalapi "session/api/server/internalapi"
)

var errorCodes = []grpcx.ErrorCode{
	{Err: app.ErrSessionNotFound, Code: codes.NotFound},
}

func NewClient(conn grpc.ClientConnInterface, timeout time.Duration) *Client {
	return &Client{
		client:  sessioninternalapi.NewSessionServiceClient(conn),
		timeout: timeout,
	}
}

type Client struct {
	client  sessioninternalapi.SessionServiceClient
	timeout time.Duration
}

func (c *Client) CreateSession(ctx context.Context, request app.CreateSessionRequest) (app.SessionGrant, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	created, err := c.client.CreateSession(ctx, &sessioninternalapi.CreateSessionRequest{
		AccountID: request.AccountID.String(),
		UserAgent: request.UserAgent,
		IPAddress: request.IPAddress,
	})
	if err != nil {
		return app.SessionGrant{}, errors.WithStack(grpcx.TranslateError(err, errorCodes...))
	}
	sessionID, err := uuid.Parse(created.SessionID)
	if err != nil {
		return app.SessionGrant{}, errors.WithStack(err)
	}
	return app.SessionGrant{
		SessionID:    sessionID,
		RefreshToken: created.RefreshToken,
		ExpiresAt:    created.ExpiresAt.AsTime(),
		NewDevice:    created.NewDevice,
	}, nil
}

func (c *Client) ValidateSession(ctx context.Context, sessionID uuid.UUID) (app.SessionValidation, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	validation, err := c.client.ValidateSession(ctx, &sessioninternalapi.ValidateSessionRequest{SessionID: sessionID.String()})
	if err != nil {
		return app.SessionValidation{}, errors.WithStack(grpcx.TranslateError(err, errorCodes...))
	}
	if !validation.Valid {
		return app.SessionValidation{}, nil
	}
	accountID, err := uuid.Parse(validation.AccountID)
	if err != nil {
		return app.SessionValidation{}, errors.WithStack(err)
	}
	return app.SessionValidation{Valid: true, AccountID: accountID}, nil
}

func (c *Client) RevokeSession(ctx context.Context, sessionID uuid.UUID, reason string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	_, err := c.client.KillSession(ctx, &sessioninternalapi.KillSessionRequest{SessionID: sessionID.String(), Reason: reason})
	return errors.WithStack(grpcx.TranslateError(err, errorCodes...))
}

func (c *Client) RevokeAccountSessions(ctx context.Context, accountID, exceptSessionID uuid.UUID, reason string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	request := &sessioninternalapi.KillAccountSessionsRequest{AccountID: accountID.String(), Reason: reason}
	if exceptSessionID != uuid.Nil() {
		request.ExceptSessionID = exceptSessionID.String()
	}
	_, err := c.client.KillAccountSessions(ctx, request)
	return errors.WithStack(grpcx.TranslateError(err, errorCodes...))
}
