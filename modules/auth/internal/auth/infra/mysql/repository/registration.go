package repository

import (
	"context"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/distributed-programming-2026/go-sdk/pkg/uow"
	"github.com/pkg/errors"

	"github.com/distributed-programming-2026/lib/common/dbuuid"
	"github.com/distributed-programming-2026/lib/common/uowx"

	"auth/internal/auth/domain/model"
	"auth/internal/auth/domain/repo"
)

func NewRegistrationRepository(unitOfWork uow.UnitOfWork) repo.RegistrationRepository {
	return &registrationRepository{
		unitOfWork: unitOfWork,
	}
}

type registrationRepository struct {
	unitOfWork uow.UnitOfWork
}

func (r *registrationRepository) Add(ctx context.Context, registration model.Registration, accountID uuid.UUID) error {
	return uowx.Exec(ctx, r.unitOfWork, func(client mysql.ClientContext) error {
		now := time.Now().UTC()
		_, err := client.ExecContext(ctx, "registration.add", `
			INSERT INTO registration_transaction (transaction_id, account_id, status, deadline, created_at, updated_at) 
			VALUES (?, ?, ?, ?, ?, ?)
		`,
			registration.TransactionID(),
			dbuuid.From(accountID),
			string(registration.Status()),
			registration.Deadline(),
			now,
			now,
		)
		return errors.WithStack(err)
	})
}

func (r *registrationRepository) FindUnfinished(ctx context.Context, limit int) ([]model.Registration, error) {
	return uowx.Query(ctx, r.unitOfWork, func(client mysql.ClientContext) ([]model.Registration, error) {
		var rows []struct {
			TransactionID string    `db:"transaction_id"`
			Status        string    `db:"status"`
			Deadline      time.Time `db:"deadline"`
		}
		err := client.SelectContext(ctx, "registration.find-unfinished", &rows, `
			SELECT transaction_id, status, deadline FROM registration_transaction
			WHERE status IN (?, ?, ?) ORDER BY updated_at LIMIT ?
		`,
			string(model.RegistrationPreparing),
			string(model.RegistrationCommitting),
			string(model.RegistrationAborting),
			limit,
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		registrations := make([]model.Registration, 0, len(rows))
		for _, row := range rows {
			registrations = append(registrations, model.NewRegistration(row.TransactionID, model.RegistrationStatus(row.Status), row.Deadline))
		}
		return registrations, nil
	})
}

func (r *registrationRepository) UpdateStatus(ctx context.Context, transactionID string, from, to model.RegistrationStatus) (bool, error) {
	return uowx.Query(ctx, r.unitOfWork, func(client mysql.ClientContext) (bool, error) {
		result, err := client.ExecContext(ctx, "registration.update-status", `
			UPDATE registration_transaction 
			SET 
			    status = ?, 
			    updated_at = ?
			WHERE 
			    transaction_id = ? AND status = ?
			`,
			string(to),
			time.Now().UTC(),
			transactionID,
			string(from),
		)
		if err != nil {
			return false, errors.WithStack(err)
		}
		affected, err := result.RowsAffected()
		return affected == 1, errors.WithStack(err)
	})
}
