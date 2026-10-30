package event

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	accountevents "account/api/events"
	"account/internal/account/domain/model"
	domainservice "account/internal/account/domain/service"
)

func NewDispatcher(factory *DispatcherFactory) domainservice.EventDispatcher {
	return &eventDispatcher{
		factory: factory,
	}
}

type eventDispatcher struct {
	factory *DispatcherFactory
}

func (d *eventDispatcher) Dispatch(ctx context.Context, domainEvent model.Event) error {
	message, err := toMessage(domainEvent)
	if err != nil {
		return err
	}
	dispatcher := d.factory.NewDispatcher(ctx)
	return errors.WithStack(dispatcher.Dispatch(message))
}

func toMessage(domainEvent model.Event) (proto.Message, error) {
	switch e := domainEvent.(type) {
	case model.AccountCreated:
		return &accountevents.AccountCreated{
			AccountID: e.AccountID.String(),
			Email:     e.Email,
			Name:      e.Name,
		}, nil
	default:
		return nil, errors.Errorf("unsupported domain event %s", domainEvent.EventName())
	}
}
