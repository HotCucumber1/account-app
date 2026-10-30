package model

type RevokeReason string

const (
	ReasonLogout          RevokeReason = "logout"
	ReasonPasswordChanged RevokeReason = "password_changed"
	ReasonRevoked         RevokeReason = "revoked"
)
