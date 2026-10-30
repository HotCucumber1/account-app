package account

import (
	"context"
	"time"
	"uuid"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/distributed-programming-2026/lib/common/grpcx"

	accountinternalapi "account/api/server/internalapi"
	"auth/internal/auth/app"
)

var errorCodes = []grpcx.ErrorCode{
	{Err: app.ErrAccountNotFound, Code: codes.NotFound},
	{Err: app.ErrAccountAlreadyExists, Code: codes.AlreadyExists},
	{Err: app.ErrInvalidAccount, Code: codes.InvalidArgument},
}

func NewClient(conn grpc.ClientConnInterface, timeout time.Duration) *Client {
	return &Client{
		client:  accountinternalapi.NewAccountServiceClient(conn),
		timeout: timeout,
	}
}

type Client struct {
	client  accountinternalapi.AccountServiceClient
	timeout time.Duration
}

func (c *Client) FindByEmail(ctx context.Context, email string) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	account, err := c.client.FindAccountByEmail(ctx, &accountinternalapi.FindAccountByEmailRequest{Email: email})
	if err != nil {
		return uuid.Nil(), errors.WithStack(grpcx.TranslateError(err, errorCodes...))
	}
	id, err := uuid.Parse(account.ID)
	return id, errors.WithStack(err)
}

func (c *Client) CreateAccount(ctx context.Context, email, name string) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	created, err := c.client.CreateAccount(ctx, &accountinternalapi.CreateAccountRequest{Email: email, Name: name})
	if err != nil {
		return uuid.Nil(), errors.WithStack(grpcx.TranslateError(err, errorCodes...))
	}
	id, err := uuid.Parse(created.AccountID)
	return id, errors.WithStack(err)
}
