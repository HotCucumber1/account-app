package domain

import (
	"github.com/pkg/errors"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrWeakPassword       = errors.New("password does not satisfy the policy")
	ErrCredentialNotFound = errors.New("credential not found")
)
