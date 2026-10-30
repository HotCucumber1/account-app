package service

import (
	"context"

	"account/internal/account/domain/model"
)

type EventDispatcher interface {
	Dispatch(ctx context.Context, event model.Event) error
}
