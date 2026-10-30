package repository

import (
	"context"
	"database/sql"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/distributed-programming-2026/go-sdk/pkg/uow"
	"github.com/pkg/errors"

	"github.com/distributed-programming-2026/lib/common/collection"
	"github.com/distributed-programming-2026/lib/common/dbuuid"
	"github.com/distributed-programming-2026/lib/common/sqlutil"
	"github.com/distributed-programming-2026/lib/common/uowx"

	"session/internal/session/domain"
	"session/internal/session/domain/model"
	"session/internal/session/domain/repo"
)

const selectSession = `
SELECT s.id, s.ip_address, s.created_at, s.expires_at, s.revoked_at, s.revoke_reason,
       d.id AS device_id, d.account_id, d.user_agent, d.first_seen_at
FROM session s
JOIN device d ON d.id = s.device_id
`

func NewSessionRepository(unitOfWork uow.UnitOfWork) repo.SessionRepository {
	return &sessionRepository{
		unitOfWork: unitOfWork,
	}
}

type sessionRepository struct {
	unitOfWork uow.UnitOfWork
}

type sessionRow struct {
	ID           dbuuid.UUID    `db:"id"`
	IPAddress    string         `db:"ip_address"`
	CreatedAt    time.Time      `db:"created_at"`
	ExpiresAt    time.Time      `db:"expires_at"`
	RevokedAt    sql.NullTime   `db:"revoked_at"`
	RevokeReason sql.NullString `db:"revoke_reason"`
	DeviceID     dbuuid.UUID    `db:"device_id"`
	AccountID    dbuuid.UUID    `db:"account_id"`
	UserAgent    string         `db:"user_agent"`
	FirstSeenAt  time.Time      `db:"first_seen_at"`
}

func (r *sessionRepository) NextID() uuid.UUID {
	return uuid.NewV7()
}

func (r *sessionRepository) Store(ctx context.Context, session model.Session) error {
	return uowx.Exec(ctx, r.unitOfWork, func(client mysql.ClientContext) error {
		_, err := client.ExecContext(ctx, "session.store", `
INSERT INTO session (id, device_id, ip_address, created_at, expires_at)
VALUES (?, ?, ?, ?, ?)
`, dbuuid.From(session.ID()), dbuuid.From(session.Device().ID()), session.IPAddress(),
			session.CreatedAt(), session.ExpiresAt())
		return errors.WithStack(err)
	})
}

func (r *sessionRepository) Update(ctx context.Context, session model.Session) error {
	return uowx.Exec(ctx, r.unitOfWork, func(client mysql.ClientContext) error {
		_, err := client.ExecContext(ctx, "session.update", `
UPDATE session SET revoked_at = ?, revoke_reason = ? WHERE id = ?
`, nullTime(session.RevokedAt()), nullString(string(session.RevokeReason())), dbuuid.From(session.ID()))
		return errors.WithStack(err)
	})
}

func (r *sessionRepository) FindByID(ctx context.Context, id uuid.UUID) (model.Session, error) {
	return uowx.Query(ctx, r.unitOfWork, func(client mysql.ClientContext) (model.Session, error) {
		var row sessionRow
		err := client.GetContext(ctx, "session.find-by-id", &row, selectSession+`WHERE s.id = ?`, dbuuid.From(id))
		if err != nil {
			return model.Session{}, errors.WithStack(sqlutil.NotFound(err, domain.ErrSessionNotFound))
		}
		return row.toDomain(), nil
	})
}

func (r *sessionRepository) FindActiveByAccount(ctx context.Context, accountID uuid.UUID, now time.Time) ([]model.Session, error) {
	return uowx.Query(ctx, r.unitOfWork, func(client mysql.ClientContext) ([]model.Session, error) {
		var rows []sessionRow
		err := client.SelectContext(ctx, "session.find-active-by-account", &rows, selectSession+`
			WHERE d.account_id = ? AND s.revoked_at IS NULL AND s.expires_at > ?`, dbuuid.From(accountID), now.UTC())
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return collection.Map(rows, sessionRow.toDomain), nil
	})
}

func (row sessionRow) toDomain() model.Session {
	session := model.NewSession(
		row.ID.UUID(),
		model.NewDevice(row.DeviceID.UUID(), row.AccountID.UUID(), row.UserAgent, row.FirstSeenAt),
		row.IPAddress,
		row.CreatedAt,
		row.ExpiresAt,
	)
	session.SetRevokedAt(row.RevokedAt.Time)
	session.SetRevokeReason(model.RevokeReason(row.RevokeReason.String))
	return session
}

func nullTime(value time.Time) sql.NullTime {
	return sql.NullTime{Time: value, Valid: !value.IsZero()}
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}
