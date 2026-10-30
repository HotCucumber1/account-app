package event

import (
	"context"
	"uuid"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	authevents "auth/api/events"
	"auth/internal/auth/domain/model"
	domainservice "auth/internal/auth/domain/service"
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
	case model.UserRegistered:
		return &authevents.UserRegistered{
			AccountID: e.AccountID.String(),
		}, nil
	case model.UserLoggedIn:
		return &authevents.UserLoggedIn{
			AccountID: e.AccountID.String(),
			SessionID: e.SessionID.String(),
			IPAddress: e.IPAddress,
			UserAgent: e.UserAgent,
			NewDevice: e.NewDevice,
		}, nil
	case model.UserLoginFailed:
		return &authevents.UserLoginFailed{
			Email:     e.Email,
			AccountID: optionalID(e.AccountID),
			IPAddress: e.IPAddress,
			Reason:    e.Reason,
		}, nil
	case model.UserLoggedOut:
		return &authevents.UserLoggedOut{
			AccountID: e.AccountID.String(),
			SessionID: e.SessionID.String(),
		}, nil
	case model.PasswordChanged:
		return &authevents.PasswordChanged{
			AccountID: e.AccountID.String(),
			SessionID: e.SessionID.String(),
		}, nil
	default:
		return nil, errors.Errorf("unsupported domain event %s", domainEvent.EventName())
	}
}

func optionalID(id uuid.UUID) string {
	if id == uuid.Nil() {
		return ""
	}
	return id.String()
}
