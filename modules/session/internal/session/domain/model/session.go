package model

import (
	"time"
	"uuid"
)

type Session struct {
	id           uuid.UUID
	device       Device
	ipAddress    string // TODO в девайс (а в локальной сети локальные вообще ломаются)
	createdAt    time.Time
	expiresAt    time.Time
	revokedAt    time.Time
	revokeReason RevokeReason
}

func NewSession(
	id uuid.UUID,
	device Device,
	ipAddress string,
	createdAt time.Time,
	expiresAt time.Time,
) Session {
	return Session{
		id:        id,
		device:    device,
		ipAddress: ipAddress,
		createdAt: createdAt,
		expiresAt: expiresAt,
	}
}

func (s *Session) ID() uuid.UUID {
	return s.id
}

func (s *Session) Device() *Device {
	return &s.device
}

func (s *Session) IPAddress() string {
	return s.ipAddress
}

func (s *Session) CreatedAt() time.Time {
	return s.createdAt
}

func (s *Session) ExpiresAt() time.Time {
	return s.expiresAt
}

func (s *Session) RevokedAt() time.Time {
	return s.revokedAt
}

func (s *Session) SetRevokedAt(revokedAt time.Time) {
	s.revokedAt = revokedAt
}

func (s *Session) RevokeReason() RevokeReason {
	return s.revokeReason
}

func (s *Session) SetRevokeReason(reason RevokeReason) {
	s.revokeReason = reason
}
