package query

import (
	"context"
	"strings"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/distributed-programming-2026/go-sdk/pkg/uow"
	"github.com/pkg/errors"

	"github.com/distributed-programming-2026/lib/common/dbuuid"
	"github.com/distributed-programming-2026/lib/common/sqlutil"
	"github.com/distributed-programming-2026/lib/common/uowx"

	"account/internal/account/app"
	appquery "account/internal/account/app/query"
	"account/internal/account/domain"
)

const selectAccount = `SELECT id, email, name, created_at FROM account `

func NewAccountQueryService(unitOfWork uow.UnitOfWork) appquery.AccountQueryService {
	return &accountQueryService{
		unitOfWork: unitOfWork,
	}
}

type accountQueryService struct {
	unitOfWork uow.UnitOfWork
}

type accountRow struct {
	ID        dbuuid.UUID `db:"id"`
	Email     string      `db:"email"`
	Name      string      `db:"name"`
	CreatedAt time.Time   `db:"created_at"`
}

func (s *accountQueryService) GetAccount(ctx context.Context, id uuid.UUID) (app.AccountData, error) {
	return s.findOne(ctx, "account.get", selectAccount+`WHERE id = ?`, dbuuid.From(id))
}

func (s *accountQueryService) FindAccountByEmail(ctx context.Context, email string) (app.AccountData, error) {
	return s.findOne(ctx, "account.find-by-email", selectAccount+`WHERE email = ?`, normalizeEmail(email))
}

func (s *accountQueryService) findOne(ctx context.Context, queryID, query string, args ...any) (app.AccountData, error) {
	return uowx.Query(ctx, s.unitOfWork, func(client mysql.ClientContext) (app.AccountData, error) {
		var row accountRow
		err := client.GetContext(ctx, queryID, &row, query, args...)
		if err != nil {
			return app.AccountData{}, errors.WithStack(sqlutil.NotFound(err, domain.ErrAccountNotFound))
		}
		return app.AccountData{
			ID:        row.ID.UUID(),
			Email:     row.Email,
			Name:      row.Name,
			CreatedAt: row.CreatedAt,
		}, nil
	})
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
