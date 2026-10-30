package query

import (
	"context"
	"uuid"

	"account/internal/account/app"
)

type AccountQueryService interface {
	GetAccount(ctx context.Context, id uuid.UUID) (app.AccountData, error)
	FindAccountByEmail(ctx context.Context, email string) (app.AccountData, error)
}
