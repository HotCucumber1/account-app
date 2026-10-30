package model

import (
	"time"
	"uuid"
)

type Account struct {
	id        uuid.UUID
	email     string
	name      string
	createdAt time.Time
}

func NewAccount(id uuid.UUID, email, name string, createdAt time.Time) Account {
	return Account{
		id:        id,
		email:     email,
		name:      name,
		createdAt: createdAt,
	}
}

func (a *Account) ID() uuid.UUID {
	return a.id
}

func (a *Account) Email() string {
	return a.email
}

func (a *Account) Name() string {
	return a.name
}

func (a *Account) CreatedAt() time.Time {
	return a.createdAt
}

func (a *Account) SetName(name string) {
	a.name = name
}
