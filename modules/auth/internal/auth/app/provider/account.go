package provider

import (
	"context"
	"uuid"
)

type AccountProvider interface {
	FindByEmail(ctx context.Context, email string) (uuid.UUID, error)
}
