package repo

import (
	"context"
	"uuid"

	"session/internal/session/domain/model"
)

type DeviceRepository interface {
	NextID() uuid.UUID
	Store(ctx context.Context, device model.Device) error
	Find(ctx context.Context, accountID uuid.UUID, userAgent string) (model.Device, error)
}
