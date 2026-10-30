package service

import (
	"context"
	"time"
	"uuid"

	"github.com/pkg/errors"

	"auth/internal/auth/domain"
	"auth/internal/auth/domain/model"
	"auth/internal/auth/domain/repo"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 100

	LoginFailureInvalidPassword = "invalid_password"
)

type AuthService struct {
	repository     repo.CredentialRepository
	passwordHasher PasswordHasher
	dispatcher     EventDispatcher
}

func NewAuthService(
	repository repo.CredentialRepository,
	passwordHasher PasswordHasher,
	dispatcher EventDispatcher,
) AuthService {
	return AuthService{
		repository:     repository,
		passwordHasher: passwordHasher,
		dispatcher:     dispatcher,
	}
}

func (s *AuthService) ValidatePassword(password string) error {
	return validatePassword(password)
}

func (s *AuthService) Register(ctx context.Context, accountID uuid.UUID, password string) error {
	hash, err := s.hash(password)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	err = s.repository.Store(ctx, model.NewCredential(accountID, hash, now, now))
	if err != nil {
		return err
	}
	return s.dispatcher.Dispatch(ctx, model.UserRegistered{
		AccountID: accountID,
	})
}

func (s *AuthService) Verify(ctx context.Context, accountID uuid.UUID, password string, attempt model.LoginAttempt) (bool, error) {
	credential, err := s.repository.Find(ctx, accountID)
	if err != nil && !errors.Is(err, domain.ErrCredentialNotFound) {
		return false, err
	}
	if err == nil && s.passwordHasher.Verify(credential.PasswordHash(), password) {
		return true, nil
	}
	return false, s.RecordLoginFailure(ctx, accountID, attempt, LoginFailureInvalidPassword)
}

func (s *AuthService) RecordLoginFailure(ctx context.Context, accountID uuid.UUID, attempt model.LoginAttempt, reason string) error {
	return s.dispatcher.Dispatch(ctx, model.UserLoginFailed{
		Email:     attempt.Email,
		AccountID: accountID,
		IPAddress: attempt.IPAddress,
		Reason:    reason,
	})
}

func (s *AuthService) RecordLogin(ctx context.Context, accountID, sessionID uuid.UUID, attempt model.LoginAttempt, newDevice bool) error {
	return s.dispatcher.Dispatch(ctx, model.UserLoggedIn{
		AccountID: accountID,
		SessionID: sessionID,
		IPAddress: attempt.IPAddress,
		UserAgent: attempt.UserAgent,
		NewDevice: newDevice,
	})
}

func (s *AuthService) RecordLogout(ctx context.Context, accountID, sessionID uuid.UUID) error {
	return s.dispatcher.Dispatch(ctx, model.UserLoggedOut{
		AccountID: accountID,
		SessionID: sessionID,
	})
}

func (s *AuthService) ChangePassword(ctx context.Context, accountID, sessionID uuid.UUID, oldPassword, newPassword string) error {
	credential, err := s.repository.Find(ctx, accountID)
	if errors.Is(err, domain.ErrCredentialNotFound) {
		return errors.WithStack(domain.ErrInvalidCredentials)
	}
	if err != nil {
		return err
	}
	if !s.passwordHasher.Verify(credential.PasswordHash(), oldPassword) {
		return errors.WithStack(domain.ErrInvalidCredentials)
	}
	hash, err := s.hash(newPassword)
	if err != nil {
		return err
	}
	credential.SetPasswordHash(hash)
	credential.SetUpdatedAt(time.Now().UTC())
	err = s.repository.Update(ctx, credential)
	if err != nil {
		return err
	}
	return s.dispatcher.Dispatch(ctx, model.PasswordChanged{
		AccountID: accountID,
		SessionID: sessionID,
	})
}

func (s *AuthService) hash(password string) (string, error) {
	err := validatePassword(password)
	if err != nil {
		return "", err
	}
	hash, err := s.passwordHasher.Hash(password)
	return hash, errors.WithStack(err)
}

func validatePassword(password string) error {
	if len(password) < minPasswordLength || len(password) > maxPasswordLength {
		return errors.WithStack(domain.ErrWeakPassword)
	}
	return nil
}
