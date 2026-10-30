package repo

import (
	"context"
	"time"
	"uuid"

	"session/internal/session/domain/model"
)

type SessionRepository interface {
	NextID() uuid.UUID
	Store(ctx context.Context, session model.Session) error
	Update(ctx context.Context, session model.Session) error
	FindByID(ctx context.Context, id uuid.UUID) (model.Session, error)
	FindActiveByAccount(ctx context.Context, accountID uuid.UUID, now time.Time) ([]model.Session, error)
}
