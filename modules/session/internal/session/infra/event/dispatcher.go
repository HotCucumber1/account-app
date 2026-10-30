package event

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	sessionevents "session/api/events"
	"session/internal/session/domain/model"
	domainservice "session/internal/session/domain/service"
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
	case model.SessionCreated:
		return &sessionevents.SessionCreated{
			SessionID: e.SessionID.String(),
			AccountID: e.AccountID.String(),
			DeviceID:  e.DeviceID.String(),
			IPAddress: e.IPAddress,
			NewDevice: e.NewDevice,
		}, nil
	case model.SessionClosed:
		return &sessionevents.SessionClosed{
			SessionID: e.SessionID.String(),
			AccountID: e.AccountID.String(),
			Reason:    string(e.Reason),
		}, nil
	default:
		return nil, errors.Errorf("unsupported domain event %s", domainEvent.EventName())
	}
}
