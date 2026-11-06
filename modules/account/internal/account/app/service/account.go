package service

import (
	"context"

	"account/internal/account/app"
	"account/internal/account/domain/model"
	domainservice "account/internal/account/domain/service"
)

type TransactionBranches interface {
	Prepare(ctx context.Context, transactionID string, action func(ctx context.Context) error) error
	Commit(ctx context.Context, transactionID string) error
	Rollback(ctx context.Context, transactionID string) error
}

type AccountService struct {
	unitOfWork UnitOfWork
	branches   TransactionBranches
	service    *domainservice.AccountService
}

func NewAccountService(unitOfWork UnitOfWork, branches TransactionBranches, service *domainservice.AccountService) *AccountService {
	return &AccountService{
		unitOfWork: unitOfWork,
		branches:   branches,
		service:    service,
	}
}

func (s *AccountService) CreateAccount(ctx context.Context, input app.CreateAccountInput) (app.AccountData, error) {
	var account model.Account
	err := s.unitOfWork.Execute(ctx, func(ctx context.Context) error {
		var err error
		account, err = s.service.Create(ctx, input.Email, input.Name)
		return err
	})
	if err != nil {
		return app.AccountData{}, err
	}
	return toAccountData(account), nil
}

func (s *AccountService) PrepareCreateAccount(ctx context.Context, input app.PrepareCreateAccountInput) error {
	return s.branches.Prepare(ctx, input.TransactionID, func(ctx context.Context) error {
		_, err := s.service.CreateWithID(ctx, input.AccountID, input.Email, input.Name)
		return err
	})
}

func (s *AccountService) CommitTransaction(ctx context.Context, transactionID string) error {
	return s.branches.Commit(ctx, transactionID)
}

func (s *AccountService) RollbackTransaction(ctx context.Context, transactionID string) error {
	return s.branches.Rollback(ctx, transactionID)
}

func (s *AccountService) UpdateAccount(ctx context.Context, input app.RenameAccountInput) (app.AccountData, error) {
	var account model.Account
	err := s.unitOfWork.Execute(ctx, func(ctx context.Context) error {
		var err error
		account, err = s.service.Update(ctx, input.AccountID, input.Name)
		return err
	})
	if err != nil {
		return app.AccountData{}, err
	}
	return toAccountData(account), nil
}

func toAccountData(account model.Account) app.AccountData {
	return app.AccountData{
		ID:        account.ID(),
		Email:     account.Email(),
		Name:      account.Name(),
		CreatedAt: account.CreatedAt(),
	}
}
