package service

import (
	"context"
	"time"
)

type UnitOfWork interface {
	Execute(ctx context.Context, action func(ctx context.Context) error) error
	ExecuteWithLock(ctx context.Context, lock string, timeout time.Duration, action func(ctx context.Context) error) error
}
