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

	"auth/internal/auth/domain"
	"auth/internal/auth/domain/model"
	"auth/internal/auth/domain/repo"
)

func NewCredentialRepository(unitOfWork uow.UnitOfWork) repo.CredentialRepository {
	return &credentialRepository{
		unitOfWork: unitOfWork,
	}
}

type credentialRepository struct {
	unitOfWork uow.UnitOfWork
}

type credentialRow struct {
	AccountID    dbuuid.UUID `db:"account_id"`
	PasswordHash string      `db:"password_hash"`
	CreatedAt    time.Time   `db:"created_at"`
	UpdatedAt    time.Time   `db:"updated_at"`
}

func (r *credentialRepository) Store(ctx context.Context, credential model.Credential) error {
	return uowx.Exec(ctx, r.unitOfWork, func(client mysql.ClientContext) error {
		_, err := client.ExecContext(ctx, "credential.store", `
			INSERT INTO credential (account_id, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?)
			`,
			dbuuid.From(credential.AccountID()),
			credential.PasswordHash(),
			credential.CreatedAt(),
			credential.UpdatedAt(),
		)
		return errors.WithStack(err)
	})
}

func (r *credentialRepository) Update(ctx context.Context, credential model.Credential) error {
	return uowx.Exec(ctx, r.unitOfWork, func(client mysql.ClientContext) error {
		_, err := client.ExecContext(ctx, "credential.update", `
UPDATE credential SET password_hash = ?, updated_at = ? WHERE account_id = ?
`, credential.PasswordHash(), credential.UpdatedAt(), dbuuid.From(credential.AccountID()))
		return errors.WithStack(err)
	})
}

func (r *credentialRepository) Find(ctx context.Context, accountID uuid.UUID) (model.Credential, error) {
	return uowx.Query(ctx, r.unitOfWork, func(client mysql.ClientContext) (model.Credential, error) {
		var row credentialRow
		err := client.GetContext(ctx, "credential.find", &row, `
SELECT account_id, password_hash, created_at, updated_at FROM credential WHERE account_id = ?
`, dbuuid.From(accountID))
		if err != nil {
			return model.Credential{}, errors.WithStack(sqlutil.NotFound(err, domain.ErrCredentialNotFound))
		}
		return model.NewCredential(row.AccountID.UUID(), row.PasswordHash, row.CreatedAt, row.UpdatedAt), nil
	})
}
