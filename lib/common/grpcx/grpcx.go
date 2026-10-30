package grpcx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type ErrorCode struct {
	Err  error
	Code codes.Code
}

func ErrorInterceptor(logger *slog.Logger, mapping ...ErrorCode) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		if err == nil {
			return resp, nil
		}
		if _, ok := status.FromError(err); ok {
			return nil, err
		}
		for _, m := range mapping {
			if errors.Is(err, m.Err) {
				return nil, status.Error(m.Code, err.Error())
			}
		}
		logger.ErrorContext(ctx, "gRPC request failed", slog.String("method", info.FullMethod), logging.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}
}

func TranslateError(err error, mapping ...ErrorCode) error {
	if err == nil {
		return nil
	}
	st, ok := status.FromError(err)
	if !ok {
		return err
	}
	for _, m := range mapping {
		if st.Code() == m.Code {
			return fmt.Errorf("%w: %s", m.Err, st.Message())
		}
	}
	return err
}

func Dial(address string) (*grpc.ClientConn, error) {
	return grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

func ParseUUID(field, value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil(), status.Errorf(codes.InvalidArgument, "%s must be a UUID", field)
	}
	return id, nil
}

func ParseUUIDOrNil(field, value string) (uuid.UUID, error) {
	if value == "" {
		return uuid.Nil(), nil
	}
	return ParseUUID(field, value)
}

type Connections []*grpc.ClientConn

func (c *Connections) Dial(address string) (*grpc.ClientConn, error) {
	conn, err := Dial(address)
	if err != nil {
		return nil, err
	}
	*c = append(*c, conn)
	return conn, nil
}

func (c Connections) Close() error {
	errs := make([]error, 0, len(c))
	for _, conn := range c {
		errs = append(errs, conn.Close())
	}
	return errors.Join(errs...)
}
