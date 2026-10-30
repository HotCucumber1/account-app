package repo

import (
	"context"
	"uuid"

	"auth/internal/auth/domain/model"
)

type CredentialRepository interface {
	Store(ctx context.Context, credential model.Credential) error
	Update(ctx context.Context, credential model.Credential) error
	Find(ctx context.Context, accountID uuid.UUID) (model.Credential, error)
}
