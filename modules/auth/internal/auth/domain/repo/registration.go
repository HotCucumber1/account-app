package repo

import (
	"context"
	"uuid"

	"auth/internal/auth/domain/model"
)

type RegistrationRepository interface {
	Add(ctx context.Context, registration model.Registration, accountID uuid.UUID) error
	FindUnfinished(ctx context.Context, limit int) ([]model.Registration, error)
	UpdateStatus(ctx context.Context, transactionID string, from, to model.RegistrationStatus) (bool, error)
}
