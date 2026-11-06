package service

import (
	"context"
	"uuid"

	"github.com/pkg/errors"

	"auth/internal/auth/app"
	appprovider "auth/internal/auth/app/provider"
	"auth/internal/auth/domain"
	"auth/internal/auth/domain/model"
	domainservice "auth/internal/auth/domain/service"
)

const (
	revokeLogout          = "logout"
	revokePasswordChanged = "password_changed"
)

type TokenIssuer interface {
	Issue(accountID, sessionID uuid.UUID) (app.AccessToken, error)
	Parse(token string) (app.Claims, error)
}

type Registrar interface {
	Register(ctx context.Context, input app.RegisterInput) (uuid.UUID, error)
}

type SessionService interface {
	CreateSession(ctx context.Context, request app.CreateSessionRequest) (app.SessionGrant, error)
	RevokeSession(ctx context.Context, sessionID uuid.UUID, reason string) error
	RevokeAccountSessions(ctx context.Context, accountID, exceptSessionID uuid.UUID, reason string) error
}

type AuthService struct {
	unitOfWork      UnitOfWork
	service         domainservice.AuthService
	accountProvider appprovider.AccountProvider
	registrar       Registrar
	sessionProvider appprovider.SessionProvider
	sessionService  SessionService
	tokenIssuer     TokenIssuer
}

func NewAuthService(
	unitOfWork UnitOfWork,
	service domainservice.AuthService,
	accountProvider appprovider.AccountProvider,
	registrar Registrar,
	sessionProvider appprovider.SessionProvider,
	sessionService SessionService,
	tokenIssuer TokenIssuer,
) *AuthService {
	return &AuthService{
		unitOfWork:      unitOfWork,
		service:         service,
		accountProvider: accountProvider,
		registrar:       registrar,
		sessionProvider: sessionProvider,
		sessionService:  sessionService,
		tokenIssuer:     tokenIssuer,
	}
}

func (s *AuthService) Register(ctx context.Context, input app.RegisterInput) (uuid.UUID, error) {
	err := s.service.ValidatePassword(input.Password)
	if err != nil {
		return uuid.Nil(), err
	}
	_, err = s.accountProvider.FindByEmail(ctx, input.Email)
	if err == nil {
		return uuid.Nil(), errors.WithStack(app.ErrAccountAlreadyExists)
	}
	if !errors.Is(err, app.ErrAccountNotFound) {
		return uuid.Nil(), err
	}
	return s.registrar.Register(ctx, input)
}

func (s *AuthService) Login(ctx context.Context, input app.LoginInput) (app.Tokens, error) {
	attempt := model.LoginAttempt{
		Email:     input.Email,
		IPAddress: input.IPAddress,
		UserAgent: input.UserAgent,
	}
	accountID, err := s.verify(ctx, input.Password, attempt)
	if err != nil {
		return app.Tokens{}, err
	}
	grant, err := s.sessionService.CreateSession(ctx, app.CreateSessionRequest{
		AccountID: accountID,
		UserAgent: input.UserAgent,
		IPAddress: input.IPAddress,
	})
	if err != nil {
		return app.Tokens{}, err
	}
	accessToken, err := s.tokenIssuer.Issue(accountID, grant.SessionID)
	if err != nil {
		return app.Tokens{}, err
	}
	err = s.unitOfWork.Execute(ctx, func(ctx context.Context) error {
		return s.service.RecordLogin(ctx, accountID, grant.SessionID, attempt, grant.NewDevice)
	})
	if err != nil {
		return app.Tokens{}, err
	}
	return app.Tokens{
		AccessToken:           accessToken.Token,
		AccessTokenExpiresAt:  accessToken.ExpiresAt,
		RefreshToken:          grant.RefreshToken,
		RefreshTokenExpiresAt: grant.ExpiresAt,
		AccountID:             accountID,
		SessionID:             grant.SessionID,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, input app.LogoutInput) error {
	err := s.requireSession(ctx, input.AccountID, input.SessionID)
	if err != nil {
		return err
	}
	err = s.sessionService.RevokeSession(ctx, input.SessionID, revokeLogout)
	if err != nil {
		return err
	}
	return s.unitOfWork.Execute(ctx, func(ctx context.Context) error {
		return s.service.RecordLogout(ctx, input.AccountID, input.SessionID)
	})
}

func (s *AuthService) ValidateToken(token string) (app.Claims, error) {
	return s.tokenIssuer.Parse(token)
}

func (s *AuthService) AuthenticateRequest(ctx context.Context, token string) (app.Claims, error) {
	claims, err := s.tokenIssuer.Parse(token)
	if err != nil {
		return app.Claims{}, errors.WithStack(err)
	}
	err = s.requireSession(ctx, claims.AccountID, claims.SessionID)
	if err != nil {
		return app.Claims{}, err
	}
	return claims, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, input app.ChangePasswordInput) error {
	err := s.requireSession(ctx, input.AccountID, input.SessionID)
	if err != nil {
		return err
	}
	err = s.unitOfWork.Execute(ctx, func(ctx context.Context) error {
		return s.service.ChangePassword(ctx, input.AccountID, input.SessionID, input.OldPassword, input.NewPassword)
	})
	if err != nil {
		return err
	}
	return s.sessionService.RevokeAccountSessions(ctx, input.AccountID, input.SessionID, revokePasswordChanged)
}

func (s *AuthService) verify(ctx context.Context, password string, attempt model.LoginAttempt) (uuid.UUID, error) {
	accountID, err := s.accountProvider.FindByEmail(ctx, attempt.Email)
	if errors.Is(err, app.ErrAccountNotFound) {
		err = s.unitOfWork.Execute(ctx, func(ctx context.Context) error {
			return s.service.RecordLoginFailure(ctx, uuid.Nil(), attempt, "unknown_account")
		})
		if err != nil {
			return uuid.Nil(), err
		}
		return uuid.Nil(), errors.WithStack(domain.ErrInvalidCredentials)
	}
	if err != nil {
		return uuid.Nil(), err
	}
	var valid bool
	err = s.unitOfWork.Execute(ctx, func(ctx context.Context) error {
		valid, err = s.service.Verify(ctx, accountID, password, attempt)
		return err
	})
	if err != nil {
		return uuid.Nil(), err
	}
	if !valid {
		return uuid.Nil(), errors.WithStack(domain.ErrInvalidCredentials)
	}
	return accountID, nil
}

func (s *AuthService) requireSession(ctx context.Context, accountID, sessionID uuid.UUID) error {
	validation, err := s.sessionProvider.ValidateSession(ctx, sessionID)
	if err != nil {
		return err
	}
	if !validation.Valid || validation.AccountID != accountID {
		return errors.WithStack(app.ErrSessionNotFound)
	}
	return nil
}
