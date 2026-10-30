package domain

import (
	"github.com/pkg/errors"
)

var (
	ErrAccountNotFound      = errors.New("account not found")
	ErrAccountAlreadyExists = errors.New("account with this email already exists")
	ErrInvalidAccount       = errors.New("invalid account data")
)
