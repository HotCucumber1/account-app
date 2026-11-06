package service

import (
	"context"

	"auth/internal/auth/app"
	domainservice "auth/internal/auth/domain/service"
)

type TransactionBranches interface {
	Prepare(ctx context.Context, transactionID string, action func(ctx context.Context) error) error
	Commit(ctx context.Context, transactionID string) error
	Rollback(ctx context.Context, transactionID string) error
}

type CredentialParticipant struct {
	branches TransactionBranches
	service  domainservice.AuthService
}

func NewCredentialParticipant(branches TransactionBranches, service domainservice.AuthService) *CredentialParticipant {
	return &CredentialParticipant{
		branches: branches,
		service:  service,
	}
}

func (p *CredentialParticipant) Name() string {
	return "auth"
}

func (p *CredentialParticipant) Prepare(ctx context.Context, request app.RegistrationRequest) error {
	return p.branches.Prepare(ctx, request.TransactionID, func(ctx context.Context) error {
		return p.service.Register(ctx, request.AccountID, request.Password)
	})
}

func (p *CredentialParticipant) Commit(ctx context.Context, transactionID string) error {
	return p.branches.Commit(ctx, transactionID)
}

func (p *CredentialParticipant) Rollback(ctx context.Context, transactionID string) error {
	return p.branches.Rollback(ctx, transactionID)
}
