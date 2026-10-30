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
)

func ErrorCodes() []grpcx.ErrorCode {
	return []grpcx.ErrorCode{
		{Err: domain.ErrAccountNotFound, Code: codes.NotFound},
		{Err: domain.ErrAccountAlreadyExists, Code: codes.AlreadyExists},
		{Err: domain.ErrInvalidAccount, Code: codes.InvalidArgument},
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

func toAccountInfo(account app.AccountData) *accountinternalapi.AccountInfo {
	return &accountinternalapi.AccountInfo{
		ID:        account.ID.String(),
		Email:     account.Email,
		Name:      account.Name,
		CreatedAt: timestamppb.New(account.CreatedAt),
	}
}
