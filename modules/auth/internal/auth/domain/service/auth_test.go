package service

import (
	"context"
	"strings"
	"testing"
	"uuid"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"auth/internal/auth/domain"
	"auth/internal/auth/domain/model"
)

type plainHasher struct{}

func (plainHasher) Hash(password string) (string, error) { return "hash:" + password, nil }
func (plainHasher) Verify(hash, password string) bool    { return hash == "hash:"+password }

type memoryRepository struct {
	credentials map[uuid.UUID]model.Credential
}

func (r *memoryRepository) Store(_ context.Context, credential model.Credential) error {
	r.credentials[credential.AccountID()] = credential
	return nil
}

func (r *memoryRepository) Update(ctx context.Context, credential model.Credential) error {
	return r.Store(ctx, credential)
}

func (r *memoryRepository) Find(_ context.Context, accountID uuid.UUID) (model.Credential, error) {
	credential, ok := r.credentials[accountID]
	if !ok {
		return model.Credential{}, errors.WithStack(domain.ErrCredentialNotFound)
	}
	return credential, nil
}

type memoryDispatcher struct {
	events []model.Event
}

func (d *memoryDispatcher) Dispatch(_ context.Context, event model.Event) error {
	d.events = append(d.events, event)
	return nil
}

func TestCredentialService(t *testing.T) {
	// Arrange
	dispatcher := &memoryDispatcher{}
	service := NewAuthService(&memoryRepository{credentials: map[uuid.UUID]model.Credential{}}, plainHasher{}, dispatcher)
	accountID := uuid.New()

	// Act
	require.ErrorIs(t, service.Register(t.Context(), accountID, "short"), domain.ErrWeakPassword)
	// Act
	require.ErrorIs(t, service.Register(t.Context(), accountID, strings.Repeat("a", 101)), domain.ErrWeakPassword)
	// Act
	require.NoError(t, service.Register(t.Context(), accountID, "password"))

	// Act
	valid, err := service.Verify(t.Context(), accountID, "password", model.LoginAttempt{})
	// Assert
	require.NoError(t, err)
	require.True(t, valid)

	// Act
	valid, err = service.Verify(t.Context(), accountID, "wrong-password", model.LoginAttempt{})
	// Assert
	require.NoError(t, err)
	require.False(t, valid)

	// Act
	require.ErrorIs(t, service.ChangePassword(t.Context(), accountID, uuid.New(), "wrong-password", "new-password"), domain.ErrInvalidCredentials)
	// Act
	require.NoError(t, service.ChangePassword(t.Context(), accountID, uuid.New(), "password", "new-password"))
	// Act
	valid, err = service.Verify(t.Context(), accountID, "new-password", model.LoginAttempt{})
	// Assert
	require.NoError(t, err)
	require.True(t, valid)
	require.Len(t, dispatcher.events, 3)
}
