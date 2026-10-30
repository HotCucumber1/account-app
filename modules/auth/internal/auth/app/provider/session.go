package provider

import (
	"context"
	"uuid"

	"auth/internal/auth/app"
)

type SessionProvider interface {
	ValidateSession(ctx context.Context, sessionID uuid.UUID) (app.SessionValidation, error)
}
