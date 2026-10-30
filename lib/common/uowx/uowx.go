package uowx

import (
	"context"
	"time"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/distributed-programming-2026/go-sdk/pkg/uow"
)

type Unit = uow.UnitOfWorkWithRepositoryProvider[mysql.ClientContext]

func NewUnit(pool mysql.ConnectionPool) Unit {
	return uow.NewUnitOfWork(pool, func(client mysql.ClientContext) mysql.ClientContext {
		return client
	})
}

type UnitOfWork struct {
	unit     uow.UnitOfWork
	lockable uow.LockableUnitOfWork
}

func NewUnitOfWork(unit Unit) *UnitOfWork {
	return &UnitOfWork{
		unit:     unit,
		lockable: uow.NewLockableUnitOfWork(unit),
	}
}

func (u *UnitOfWork) Execute(ctx context.Context, action func(ctx context.Context) error) error {
	return u.unit.ExecuteWithClientContext(ctx, func(mysql.ClientContext) error {
		return action(ctx)
	})
}

func (u *UnitOfWork) ExecuteWithLock(ctx context.Context, lock string, timeout time.Duration, action func(ctx context.Context) error) error {
	return u.lockable.ExecuteWithClientContext(ctx, []uow.LockOptions{{Name: lock, Timeout: timeout}}, func(mysql.ClientContext) error {
		return action(ctx)
	})
}

func Exec(ctx context.Context, unit uow.UnitOfWork, f func(client mysql.ClientContext) error) error {
	_, err := Query(ctx, unit, func(client mysql.ClientContext) (struct{}, error) {
		return struct{}{}, f(client)
	})
	return err
}

func Query[T any](ctx context.Context, unit uow.UnitOfWork, f func(client mysql.ClientContext) (T, error)) (T, error) {
	var (
		result T
		err    error
	)
	unitErr := unit.ExecuteWithClientContext(ctx, func(client mysql.ClientContext) error {
		result, err = f(client)
		return nil
	})
	if unitErr != nil {
		var zero T
		return zero, unitErr
	}
	return result, err
}
