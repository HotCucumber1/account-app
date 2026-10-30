package domain

import (
	"github.com/pkg/errors"
)

var (
	ErrSessionNotFound      = errors.New("session not found")
	ErrInvalidRevokeReason  = errors.New("invalid revoke reason")
	ErrInvalidSessionParams = errors.New("invalid session parameters")
	ErrDeviceNotFound       = errors.New("device not found")
	ErrDeviceAlreadyExists  = errors.New("device already exists")
)
