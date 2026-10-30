package repository

import (
	"context"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/distributed-programming-2026/go-sdk/pkg/uow"
	"github.com/pkg/errors"

	"github.com/distributed-programming-2026/lib/common/dbuuid"
	"github.com/distributed-programming-2026/lib/common/secret"
	"github.com/distributed-programming-2026/lib/common/sqlutil"
	"github.com/distributed-programming-2026/lib/common/uowx"

	"session/internal/session/domain"
	"session/internal/session/domain/model"
	"session/internal/session/domain/repo"
)

func NewDeviceRepository(unitOfWork uow.UnitOfWork) repo.DeviceRepository {
	return &deviceRepository{
		unitOfWork: unitOfWork,
	}
}

type deviceRepository struct {
	unitOfWork uow.UnitOfWork
}

type deviceRow struct {
	ID          dbuuid.UUID `db:"id"`
	AccountID   dbuuid.UUID `db:"account_id"`
	UserAgent   string      `db:"user_agent"`
	FirstSeenAt time.Time   `db:"first_seen_at"`
}

func (r *deviceRepository) NextID() uuid.UUID {
	return uuid.NewV7()
}

func (r *deviceRepository) Store(ctx context.Context, device model.Device) error {
	return uowx.Exec(ctx, r.unitOfWork, func(client mysql.ClientContext) error {
		_, err := client.ExecContext(ctx, "device.store", `
			INSERT INTO device (id, account_id, user_agent_sha256, user_agent, first_seen_at) VALUES (?, ?, ?, ?, ?)`,
			dbuuid.From(device.ID()),
			dbuuid.From(device.AccountID()),
			secret.Digest(device.UserAgent()),
			device.UserAgent(),
			device.FirstSeenAt(),
		)
		if sqlutil.IsDuplicate(err) {
			return errors.WithStack(domain.ErrDeviceAlreadyExists)
		}
		return errors.WithStack(err)
	})
}

func (r *deviceRepository) Find(ctx context.Context, accountID uuid.UUID, userAgent string) (model.Device, error) {
	return uowx.Query(ctx, r.unitOfWork, func(client mysql.ClientContext) (model.Device, error) {
		var row deviceRow
		err := client.GetContext(ctx, "device.find", &row, `
			SELECT id, account_id, user_agent, first_seen_at
			FROM device
			WHERE account_id = ? AND user_agent_sha256 = ?
			`,
			dbuuid.From(accountID),
			secret.Digest(userAgent),
		)
		if err != nil {
			return model.Device{}, errors.WithStack(sqlutil.NotFound(err, domain.ErrDeviceNotFound))
		}
		return model.NewDevice(row.ID.UUID(), row.AccountID.UUID(), row.UserAgent, row.FirstSeenAt), nil
	})
}
