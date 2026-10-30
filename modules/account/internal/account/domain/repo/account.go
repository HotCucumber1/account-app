package repo

import (
	"context"
	"uuid"

	"account/internal/account/domain/model"
)

type AccountRepository interface {
	NextID() uuid.UUID
	Store(ctx context.Context, account model.Account) error
	FindByID(ctx context.Context, id uuid.UUID) (model.Account, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}
