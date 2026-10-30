package repository

import (
	"context"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/distributed-programming-2026/go-sdk/pkg/uow"
	"github.com/pkg/errors"

	"github.com/distributed-programming-2026/lib/common/dbuuid"
	"github.com/distributed-programming-2026/lib/common/sqlutil"
	"github.com/distributed-programming-2026/lib/common/uowx"

	"account/internal/account/domain"
	"account/internal/account/domain/model"
	"account/internal/account/domain/repo"
)

func NewAccountRepository(unitOfWork uow.UnitOfWork) repo.AccountRepository {
	return &accountRepository{
		unitOfWork: unitOfWork,
	}
}

type accountRepository struct {
	unitOfWork uow.UnitOfWork
}

func (r *accountRepository) NextID() uuid.UUID {
	return uuid.NewV7()
}

func (r *accountRepository) Store(ctx context.Context, account model.Account) error {
	return uowx.Exec(ctx, r.unitOfWork, func(client mysql.ClientContext) error {
		_, err := client.ExecContext(ctx, "account.store", `
INSERT INTO account (id, email, name, created_at) VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE name = VALUES(name)
`, dbuuid.From(account.ID()), account.Email(), account.Name(), account.CreatedAt())
		return errors.WithStack(err)
	})
}

func (r *accountRepository) FindByID(ctx context.Context, id uuid.UUID) (model.Account, error) {
	return uowx.Query(ctx, r.unitOfWork, func(client mysql.ClientContext) (model.Account, error) {
		var row struct {
			ID        dbuuid.UUID `db:"id"`
			Email     string      `db:"email"`
			Name      string      `db:"name"`
			CreatedAt time.Time   `db:"created_at"`
		}
		err := client.GetContext(ctx, "account.find-by-id", &row, `
SELECT id, email, name, created_at FROM account WHERE id = ? FOR UPDATE
`, dbuuid.From(id))
		if err != nil {
			return model.Account{}, errors.WithStack(
				sqlutil.NotFound(err, domain.ErrAccountNotFound),
			)
		}
		return model.NewAccount(row.ID.UUID(), row.Email, row.Name, row.CreatedAt), nil
	})
}

func (r *accountRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return uowx.Query(ctx, r.unitOfWork, func(client mysql.ClientContext) (bool, error) {
		var count int
		err := client.GetContext(ctx, "account.exists-by-email", &count, `SELECT COUNT(*) FROM account WHERE email = ?`, email)
		return count > 0, errors.WithStack(err)
	})
}
