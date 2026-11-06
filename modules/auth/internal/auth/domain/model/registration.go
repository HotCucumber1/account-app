package model

import "time"

type RegistrationStatus string

const (
	RegistrationPreparing  RegistrationStatus = "PREPARING"
	RegistrationCommitting RegistrationStatus = "COMMITTING"
	RegistrationAborting   RegistrationStatus = "ABORTING"
	RegistrationCommitted  RegistrationStatus = "COMMITTED"
	RegistrationAborted    RegistrationStatus = "ABORTED"
)

type Registration struct {
	transactionID string
	status        RegistrationStatus
	deadline      time.Time
}

func NewRegistration(transactionID string, status RegistrationStatus, deadline time.Time) Registration {
	return Registration{
		transactionID: transactionID,
		status:        status,
		deadline:      deadline,
	}
}

func (r *Registration) TransactionID() string {
	return r.transactionID
}

func (r *Registration) Status() RegistrationStatus {
	return r.status
}

func (r *Registration) Deadline() time.Time {
	return r.deadline
}
