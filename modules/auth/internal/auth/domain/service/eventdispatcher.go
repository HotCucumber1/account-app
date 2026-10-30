package service

import (
	"context"

	"auth/internal/auth/domain/model"
)

type EventDispatcher interface {
	Dispatch(ctx context.Context, event model.Event) error
}
