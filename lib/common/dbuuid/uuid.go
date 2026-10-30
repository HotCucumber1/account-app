package dbuuid

import (
	"database/sql/driver"
	"fmt"
	"uuid"
)

type UUID uuid.UUID

func From(id uuid.UUID) UUID {
	return UUID(id)
}

func (u UUID) UUID() uuid.UUID {
	return uuid.UUID(u)
}

func (u UUID) Value() (driver.Value, error) {
	return u[:], nil
}

func (u *UUID) Scan(src any) error {
	raw, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("dbuuid: unsupported source type %T", src)
	}
	if len(raw) != len(u) {
		return fmt.Errorf("dbuuid: invalid length %d", len(raw))
	}
	copy(u[:], raw)
	return nil
}

type NullUUID struct {
	UUID  UUID
	Valid bool
}

func (n *NullUUID) UUIDOrNil() uuid.UUID {
	if !n.Valid {
		return uuid.Nil()
	}
	return n.UUID.UUID()
}

func (n *NullUUID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.UUID.Value()
}

func (n *NullUUID) Scan(src any) error {
	if src == nil {
		n.UUID, n.Valid = UUID{}, false
		return nil
	}
	n.Valid = true
	return n.UUID.Scan(src)
}
