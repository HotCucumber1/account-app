package service

import (
	"context"

	"session/internal/session/domain/model"
)

type EventDispatcher interface {
	Dispatch(ctx context.Context, event model.Event) error
}
