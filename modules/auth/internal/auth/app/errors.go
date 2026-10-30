package app

import (
	"github.com/pkg/errors"
)

var (
	ErrAccountNotFound      = errors.New("account not found")
	ErrAccountAlreadyExists = errors.New("account already exists")
	ErrInvalidAccount       = errors.New("invalid account data")
	ErrInvalidToken         = errors.New("access token is invalid")
	ErrSessionNotFound      = errors.New("session not found")
)
