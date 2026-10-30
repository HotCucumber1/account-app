package model

import "uuid"

type Event interface {
	EventName() string
}

type SessionCreated struct {
	SessionID uuid.UUID
	AccountID uuid.UUID
	DeviceID  uuid.UUID
	IPAddress string
	NewDevice bool
}

func (SessionCreated) EventName() string {
	return "SessionCreated"
}

type SessionClosed struct {
	SessionID uuid.UUID
	AccountID uuid.UUID
	Reason    RevokeReason
}

func (SessionClosed) EventName() string {
	return "SessionClosed"
}
