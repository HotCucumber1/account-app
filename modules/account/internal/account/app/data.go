package app

import (
	"time"
	"uuid"
)

type CreateAccountInput struct {
	Email string
	Name  string
}

type RenameAccountInput struct {
	AccountID uuid.UUID
	Name      string
}

type AccountData struct {
	ID        uuid.UUID
	Email     string
	Name      string
	CreatedAt time.Time
}

type PrepareCreateAccountInput struct {
	TransactionID string
	AccountID     uuid.UUID
	Email         string
	Name          string
}
