package model

import "uuid"

type Event interface {
	EventName() string
}

type AccountCreated struct {
	AccountID uuid.UUID
	Email     string
	Name      string
}

func (AccountCreated) EventName() string {
	return "AccountCreated"
}
