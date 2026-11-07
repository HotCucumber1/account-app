package internalapi

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/distributed-programming-2026/lib/common/grpcx"

	accountinternalapi "account/api/server/internalapi"
	"account/internal/account/app"
	appquery "account/internal/account/app/query"
	appservice "account/internal/account/app/service"
	"account/internal/account/domain"

	"github.com/distributed-programming-2026/lib/common/xa"
)

func ErrorCodes() []grpcx.ErrorCode {
	return []grpcx.ErrorCode{
		{Err: domain.ErrAccountNotFound, Code: codes.NotFound},
		{Err: domain.ErrAccountAlreadyExists, Code: codes.AlreadyExists},
		{Err: domain.ErrInvalidAccount, Code: codes.InvalidArgument},
		{Err: xa.ErrInvalidTransactionID, Code: codes.InvalidArgument},
	}
}

func NewHandler(service *appservice.AccountService, queryService appquery.AccountQueryService) accountinternalapi.AccountServiceServer {
	return &handler{
		service:      service,
		queryService: queryService,
	}
}

type handler struct {
	service      *appservice.AccountService
	queryService appquery.AccountQueryService
	accountinternalapi.UnimplementedAccountServiceServer
}

func (h handler) CreateAccount(ctx context.Context, req *accountinternalapi.CreateAccountRequest) (*accountinternalapi.CreateAccountResponse, error) {
	account, err := h.service.CreateAccount(ctx, app.CreateAccountInput{Email: req.Email, Name: req.Name})
	if err != nil {
		return nil, err
	}
	return &accountinternalapi.CreateAccountResponse{AccountID: account.ID.String()}, nil
}

func (h handler) GetAccountInfo(ctx context.Context, req *accountinternalapi.GetAccountInfoRequest) (*accountinternalapi.AccountInfo, error) {
	id, err := grpcx.ParseUUID("AccountID", req.AccountID)
	if err != nil {
		return nil, err
	}
	account, err := h.queryService.GetAccount(ctx, id)
	if err != nil {
		return nil, err
	}
	return toAccountInfo(account), nil
}

func (h handler) FindAccountByEmail(ctx context.Context, req *accountinternalapi.FindAccountByEmailRequest) (*accountinternalapi.AccountInfo, error) {
	account, err := h.queryService.FindAccountByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	return toAccountInfo(account), nil
}

func (h handler) PrepareCreateAccount(ctx context.Context, req *accountinternalapi.PrepareCreateAccountRequest) (*accountinternalapi.PrepareCreateAccountResponse, error) {
	id, err := grpcx.ParseUUID("AccountID", req.AccountID)
	if err != nil {
		return nil, err
	}
	err = h.service.PrepareCreateAccount(ctx, app.PrepareCreateAccountInput{
		TransactionID: req.TransactionID,
		AccountID:     id,
		Email:         req.Email,
		Name:          req.Name,
	})
	if err != nil {
		return nil, err
	}
	return &accountinternalapi.PrepareCreateAccountResponse{}, nil
}

func (h handler) CommitTransaction(ctx context.Context, req *accountinternalapi.TransactionRequest) (*accountinternalapi.TransactionResponse, error) {
	return &accountinternalapi.TransactionResponse{}, h.service.CommitTransaction(ctx, req.TransactionID)
}

func (h handler) RollbackTransaction(ctx context.Context, req *accountinternalapi.TransactionRequest) (*accountinternalapi.TransactionResponse, error) {
	return &accountinternalapi.TransactionResponse{}, h.service.RollbackTransaction(ctx, req.TransactionID)
}

func toAccountInfo(account app.AccountData) *accountinternalapi.AccountInfo {
	return &accountinternalapi.AccountInfo{
		ID:        account.ID.String(),
		Email:     account.Email,
		Name:      account.Name,
		CreatedAt: timestamppb.New(account.CreatedAt),
	}
}
