package service

import (
	"context"

	"session/internal/session/app"
	domainservice "session/internal/session/domain/service"
)

type SessionService struct {
	unitOfWork UnitOfWork
	service    *domainservice.SessionService
}

func NewSessionService(unitOfWork UnitOfWork, service *domainservice.SessionService) *SessionService {
	return &SessionService{
		unitOfWork: unitOfWork,
		service:    service,
	}
}

func (s *SessionService) CreateSession(ctx context.Context, input app.CreateSessionInput) (app.CreatedSession, error) {
	var created domainservice.CreatedSession
	err := s.unitOfWork.Execute(ctx, func(ctx context.Context) error {
		var err error
		created, err = s.service.Create(ctx, input.AccountID, input.UserAgent, input.IPAddress)
		return err
	})
	if err != nil {
		return app.CreatedSession{}, err
	}
	return app.CreatedSession{
		SessionID: created.Session.ID(),
		ExpiresAt: created.Session.ExpiresAt(),
		NewDevice: created.NewDevice,
	}, nil
}

func (s *SessionService) RevokeSession(ctx context.Context, input app.KillSessionInput) error {
	return s.unitOfWork.Execute(ctx, func(ctx context.Context) error {
		return s.service.Revoke(ctx, input.SessionID, input.Reason)
	})
}

func (s *SessionService) KillAccountSessions(ctx context.Context, input app.KillAccountSessionsInput) (int, error) {
	var killed int
	err := s.unitOfWork.Execute(ctx, func(ctx context.Context) error {
		var err error
		killed, err = s.service.KillAccountSessions(ctx, input.AccountID, input.ExceptSessionID, input.Reason)
		return err
	})
	return killed, err
}
