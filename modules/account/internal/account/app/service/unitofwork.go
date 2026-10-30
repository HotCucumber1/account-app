package service

import (
	"context"
)

type UnitOfWork interface {
	Execute(ctx context.Context, action func(ctx context.Context) error) error
}
