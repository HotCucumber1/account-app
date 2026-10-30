package app

import (
	"time"
	"uuid"
)

type RegisterInput struct {
	Email    string
	Password string
	Name     string
}

type LoginInput struct {
	Email     string
	Password  string
	UserAgent string
	IPAddress string
}

type LogoutInput struct {
	AccountID uuid.UUID
	SessionID uuid.UUID
}

type ChangePasswordInput struct {
	AccountID   uuid.UUID
	SessionID   uuid.UUID
	OldPassword string
	NewPassword string
}

type Tokens struct {
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
	AccountID             uuid.UUID
	SessionID             uuid.UUID
}

type Claims struct {
	AccountID uuid.UUID
	SessionID uuid.UUID
	ExpiresAt time.Time
}

type AccessToken struct {
	Token     string
	ExpiresAt time.Time
}

type CreateSessionRequest struct {
	AccountID uuid.UUID
	UserAgent string
	IPAddress string
}

type SessionGrant struct {
	SessionID    uuid.UUID
	RefreshToken string
	ExpiresAt    time.Time
	NewDevice    bool
}

type SessionValidation struct {
	Valid     bool
	AccountID uuid.UUID
}
