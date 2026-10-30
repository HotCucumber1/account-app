package model

import (
	"time"
	"uuid"
)

type Device struct {
	id          uuid.UUID
	accountID   uuid.UUID // todo в сессию
	userAgent   string    // todo если он поменялся? (а еще и тестик сразу на это)
	firstSeenAt time.Time
}

func NewDevice(id, accountID uuid.UUID, userAgent string, firstSeenAt time.Time) Device {
	return Device{
		id:          id,
		accountID:   accountID,
		userAgent:   userAgent,
		firstSeenAt: firstSeenAt,
	}
}

func (d *Device) ID() uuid.UUID {
	return d.id
}

func (d *Device) AccountID() uuid.UUID {
	return d.accountID
}

func (d *Device) UserAgent() string {
	return d.userAgent
}

func (d *Device) FirstSeenAt() time.Time {
	return d.firstSeenAt
}
