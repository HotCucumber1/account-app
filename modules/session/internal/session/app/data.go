package app

import (
	"time"
	"uuid"
)

type CreateSessionInput struct {
	AccountID uuid.UUID
	UserAgent string
	IPAddress string
}

type CreatedSession struct {
	SessionID uuid.UUID
	ExpiresAt time.Time
	NewDevice bool
}

type KillSessionInput struct {
	SessionID uuid.UUID
	Reason    string
}

type KillAccountSessionsInput struct {
	AccountID       uuid.UUID
	ExceptSessionID uuid.UUID
	Reason          string
}

type Session struct {
	ID           uuid.UUID
	AccountID    uuid.UUID
	DeviceID     uuid.UUID
	UserAgent    string
	IPAddress    string
	Status       string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	RevokeReason string
}

type Validation struct {
	Valid     bool
	AccountID uuid.UUID
	ExpiresAt time.Time
}
