package model

import (
	"time"
	"uuid"
)

type Credential struct {
	accountID    uuid.UUID
	passwordHash string
	createdAt    time.Time
	updatedAt    time.Time
}

func NewCredential(accountID uuid.UUID, passwordHash string, createdAt, updatedAt time.Time) Credential {
	return Credential{
		accountID:    accountID,
		passwordHash: passwordHash,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

func (c *Credential) AccountID() uuid.UUID {
	return c.accountID
}

func (c *Credential) PasswordHash() string {
	return c.passwordHash
}

func (c *Credential) SetPasswordHash(passwordHash string) {
	c.passwordHash = passwordHash
}

func (c *Credential) CreatedAt() time.Time {
	return c.createdAt
}

func (c *Credential) UpdatedAt() time.Time {
	return c.updatedAt
}

func (c *Credential) SetUpdatedAt(updatedAt time.Time) {
	c.updatedAt = updatedAt
}
