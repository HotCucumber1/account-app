package internalapi

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/distributed-programming-2026/lib/common/grpcx"

	authinternalapi "auth/api/server/internalapi"
	"auth/internal/auth/app"
	appservice "auth/internal/auth/app/service"
	"auth/internal/auth/domain"
)

func ErrorCodes() []grpcx.ErrorCode {
	return []grpcx.ErrorCode{
		{Err: domain.ErrInvalidCredentials, Code: codes.Unauthenticated},
		{Err: app.ErrInvalidToken, Code: codes.Unauthenticated},
		{Err: app.ErrSessionNotFound, Code: codes.Unauthenticated},
		{Err: domain.ErrWeakPassword, Code: codes.InvalidArgument},
		{Err: app.ErrInvalidAccount, Code: codes.InvalidArgument},
		{Err: app.ErrAccountAlreadyExists, Code: codes.AlreadyExists},
	}
}

func NewHandler(service *appservice.AuthService) authinternalapi.AuthServiceServer {
	return &handler{
		service: service,
	}
}

type handler struct {
	service *appservice.AuthService
	authinternalapi.UnimplementedAuthServiceServer
}

func (h handler) Register(ctx context.Context, req *authinternalapi.RegisterRequest) (*authinternalapi.RegisterResponse, error) {
	accountID, err := h.service.Register(ctx, app.RegisterInput{Email: req.Email, Password: req.Password, Name: req.Name})
	if err != nil {
		return nil, err
	}
	return &authinternalapi.RegisterResponse{AccountID: accountID.String()}, nil
}

func (h handler) Authenticate(ctx context.Context, req *authinternalapi.AuthenticateRequest) (*authinternalapi.Tokens, error) {
	tokens, err := h.service.Login(ctx, app.LoginInput{
		Email:     req.Email,
		Password:  req.Password,
		UserAgent: req.UserAgent,
		IPAddress: req.IPAddress,
	})
	if err != nil {
		return nil, err
	}
	return &authinternalapi.Tokens{
		AccessToken:           tokens.AccessToken,
		AccessTokenExpiresAt:  timestamppb.New(tokens.AccessTokenExpiresAt),
		RefreshToken:          tokens.RefreshToken,
		RefreshTokenExpiresAt: timestamppb.New(tokens.RefreshTokenExpiresAt),
		AccountID:             tokens.AccountID.String(),
		SessionID:             tokens.SessionID.String(),
	}, nil
}

func (h handler) RevokeToken(ctx context.Context, req *authinternalapi.RevokeTokenRequest) (*authinternalapi.RevokeTokenResponse, error) {
	accountID, err := grpcx.ParseUUID("AccountID", req.AccountID)
	if err != nil {
		return nil, err
	}
	sessionID, err := grpcx.ParseUUID("SessionID", req.SessionID)
	if err != nil {
		return nil, err
	}
	return &authinternalapi.RevokeTokenResponse{}, h.service.Logout(ctx, app.LogoutInput{AccountID: accountID, SessionID: sessionID})
}

func (h handler) ValidateToken(_ context.Context, req *authinternalapi.ValidateTokenRequest) (*authinternalapi.ValidateTokenResponse, error) {
	claims, err := h.service.ValidateToken(req.AccessToken)
	if err != nil {
		return nil, err
	}
	return &authinternalapi.ValidateTokenResponse{
		AccountID: claims.AccountID.String(),
		SessionID: claims.SessionID.String(),
		ExpiresAt: timestamppb.New(claims.ExpiresAt),
	}, nil
}

func (h handler) ChangePassword(ctx context.Context, req *authinternalapi.ChangePasswordRequest) (*authinternalapi.ChangePasswordResponse, error) {
	accountID, err := grpcx.ParseUUID("AccountID", req.AccountID)
	if err != nil {
		return nil, err
	}
	sessionID, err := grpcx.ParseUUID("SessionID", req.SessionID)
	if err != nil {
		return nil, err
	}
	return &authinternalapi.ChangePasswordResponse{}, h.service.ChangePassword(ctx, app.ChangePasswordInput{
		AccountID:   accountID,
		SessionID:   sessionID,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	})
}
