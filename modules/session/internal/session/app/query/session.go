package query

import (
	"context"
	"uuid"

	"session/internal/session/app"
)

type SessionQueryService interface {
	GetSession(ctx context.Context, id uuid.UUID) (app.Session, error)
	ValidateSession(ctx context.Context, id uuid.UUID) (app.Validation, error)
}
