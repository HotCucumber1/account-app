package query

import (
	"context"
	"database/sql"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/distributed-programming-2026/go-sdk/pkg/uow"
	"github.com/pkg/errors"

	"github.com/distributed-programming-2026/lib/common/dbuuid"
	"github.com/distributed-programming-2026/lib/common/sqlutil"
	"github.com/distributed-programming-2026/lib/common/uowx"

	"session/internal/session/app"
	appquery "session/internal/session/app/query"
	"session/internal/session/domain"
)

func NewSessionQueryService(unitOfWork uow.UnitOfWork) appquery.SessionQueryService {
	return &sessionQueryService{
		unitOfWork: unitOfWork,
	}
}

type sessionQueryService struct {
	unitOfWork uow.UnitOfWork
}

type sessionRow struct {
	ID           dbuuid.UUID    `db:"id"`
	AccountID    dbuuid.UUID    `db:"account_id"`
	DeviceID     dbuuid.UUID    `db:"device_id"`
	UserAgent    string         `db:"user_agent"`
	IPAddress    string         `db:"ip_address"`
	Status       string         `db:"status"`
	CreatedAt    time.Time      `db:"created_at"`
	ExpiresAt    time.Time      `db:"expires_at"`
	RevokeReason sql.NullString `db:"revoke_reason"`
}

func (s *sessionQueryService) GetSession(ctx context.Context, id uuid.UUID) (app.Session, error) {
	return uowx.Query(ctx, s.unitOfWork, func(client mysql.ClientContext) (app.Session, error) {
		var row sessionRow
		err := client.GetContext(ctx, "session.get", &row, `
SELECT s.id, d.account_id, d.id AS device_id, d.user_agent, s.ip_address, s.created_at, s.expires_at, s.revoke_reason,
       CASE
           WHEN s.revoked_at IS NOT NULL THEN 'revoked'
           WHEN s.expires_at <= UTC_TIMESTAMP(6) THEN 'expired'
           ELSE 'active'
       END AS status
FROM session s
JOIN device d ON d.id = s.device_id
WHERE s.id = ?
`, dbuuid.From(id))
		if err != nil {
			return app.Session{}, errors.WithStack(sqlutil.NotFound(err, domain.ErrSessionNotFound))
		}
		return app.Session{
			ID:           row.ID.UUID(),
			AccountID:    row.AccountID.UUID(),
			DeviceID:     row.DeviceID.UUID(),
			UserAgent:    row.UserAgent,
			IPAddress:    row.IPAddress,
			Status:       row.Status,
			CreatedAt:    row.CreatedAt,
			ExpiresAt:    row.ExpiresAt,
			RevokeReason: row.RevokeReason.String,
		}, nil
	})
}

func (s *sessionQueryService) ValidateSession(ctx context.Context, id uuid.UUID) (app.Validation, error) {
	return uowx.Query(ctx, s.unitOfWork, func(client mysql.ClientContext) (app.Validation, error) {
		var row struct {
			AccountID dbuuid.UUID `db:"account_id"`
			ExpiresAt time.Time   `db:"expires_at"`
		}
		err := client.GetContext(ctx, "session.validate", &row, `
SELECT d.account_id, s.expires_at
FROM session s
JOIN device d ON d.id = s.device_id
WHERE s.id = ? AND s.revoked_at IS NULL AND s.expires_at > UTC_TIMESTAMP(6)
`, dbuuid.From(id))
		if errors.Is(err, sql.ErrNoRows) {
			return app.Validation{}, nil
		}
		if err != nil {
			return app.Validation{}, errors.WithStack(err)
		}
		return app.Validation{Valid: true, AccountID: row.AccountID.UUID(), ExpiresAt: row.ExpiresAt}, nil
	})
}
