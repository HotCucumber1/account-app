package service

import (
	"context"
	"strings"
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"

	"account/internal/account/domain"
	"account/internal/account/domain/model"
)

type memoryRepository struct {
	accounts map[string]model.Account
}

func (r *memoryRepository) NextID() uuid.UUID {
	return uuid.NewV7()
}

func (r *memoryRepository) Store(_ context.Context, account model.Account) error {
	r.accounts[account.Email()] = account
	return nil
}

func (r *memoryRepository) FindByID(_ context.Context, id uuid.UUID) (model.Account, error) {
	for _, account := range r.accounts {
		if account.ID() == id {
			return account, nil
		}
	}
	return model.Account{}, domain.ErrAccountNotFound
}

func (r *memoryRepository) ExistsByEmail(_ context.Context, email string) (bool, error) {
	_, ok := r.accounts[email]
	return ok, nil
}

type memoryDispatcher struct {
	events []model.Event
}

func (d *memoryDispatcher) Dispatch(_ context.Context, event model.Event) error {
	d.events = append(d.events, event)
	return nil
}

func TestCreateAccount(t *testing.T) {
	// Arrange
	dispatcher := &memoryDispatcher{}
	service := NewAccountService(&memoryRepository{accounts: map[string]model.Account{}}, dispatcher)

	// Act
	account, err := service.Create(t.Context(), "  John@Example.COM ", " John ")
	// Assert
	require.NoError(t, err)
	require.Equal(t, "john@example.com", account.Email())
	require.Equal(t, "John", account.Name())
	require.Len(t, dispatcher.events, 1)

	// Act
	_, err = service.Create(t.Context(), "john@example.com", "Twin")
	// Assert
	require.ErrorIs(t, err, domain.ErrAccountAlreadyExists)

	for _, email := range []string{"not-an-email", "John <other@example.com>"} {
		// Act
		_, err = service.Create(t.Context(), email, "John")
		// Assert
		require.ErrorIs(t, err, domain.ErrInvalidAccount)
	}
	// Act
	_, err = service.Create(t.Context(), "ann@example.com", " ")
	// Assert
	require.ErrorIs(t, err, domain.ErrInvalidAccount)
}

func TestRenameAccount(t *testing.T) {
	// Arrange
	dispatcher := &memoryDispatcher{}
	service := NewAccountService(&memoryRepository{accounts: map[string]model.Account{}}, dispatcher)
	account, err := service.Create(t.Context(), "john@example.com", "John")
	require.NoError(t, err)

	// Act
	renamed, err := service.Update(t.Context(), account.ID(), "  John Smith ")
	// Assert
	require.NoError(t, err)
	require.Equal(t, "John Smith", renamed.Name())
	require.Equal(t, "john@example.com", renamed.Email())
	require.Len(t, dispatcher.events, 1)

	for _, name := range []string{"", "   ", strings.Repeat("я", maxNameLength+1)} {
		// Act
		_, err = service.Update(t.Context(), account.ID(), name)
		// Assert
		require.ErrorIs(t, err, domain.ErrInvalidAccount)
	}

	// Act
	_, err = service.Update(t.Context(), uuid.NewV7(), "Ghost")
	// Assert
	require.ErrorIs(t, err, domain.ErrAccountNotFound)
}
