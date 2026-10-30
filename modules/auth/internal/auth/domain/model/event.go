package model

import "uuid"

type Event interface {
	EventName() string
}

type UserRegistered struct {
	AccountID uuid.UUID
}

func (UserRegistered) EventName() string {
	return "UserRegistered"
}

type UserLoggedIn struct {
	AccountID uuid.UUID
	SessionID uuid.UUID
	IPAddress string
	UserAgent string
	NewDevice bool
}

func (UserLoggedIn) EventName() string {
	return "UserLoggedIn"
}

type UserLoginFailed struct {
	Email     string
	AccountID uuid.UUID
	IPAddress string
	Reason    string
}

func (UserLoginFailed) EventName() string {
	return "UserLoginFailed"
}

type UserLoggedOut struct {
	AccountID uuid.UUID
	SessionID uuid.UUID
}

func (UserLoggedOut) EventName() string {
	return "UserLoggedOut"
}

type PasswordChanged struct {
	AccountID uuid.UUID
	SessionID uuid.UUID
}

func (PasswordChanged) EventName() string {
	return "PasswordChanged"
}
