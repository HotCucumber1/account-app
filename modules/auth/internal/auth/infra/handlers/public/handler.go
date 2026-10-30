package public

import (
	"context"
	"uuid"

	openapitypes "github.com/oapi-codegen/runtime/types"
	"github.com/pkg/errors"

	"github.com/distributed-programming-2026/lib/common/openapix"

	"auth/api/server/publicapi"
	"auth/internal/auth/app"
	appservice "auth/internal/auth/app/service"
	"auth/internal/auth/domain"
)

func NewHandler(service *appservice.AuthService) publicapi.StrictServerInterface {
	return &handler{
		service: service,
	}
}

type handler struct {
	service *appservice.AuthService
}

func (h handler) Register(ctx context.Context, req publicapi.RegisterRequestObject) (publicapi.RegisterResponseObject, error) {
	accountID, err := h.service.Register(ctx, app.RegisterInput{Email: req.Body.Email, Password: req.Body.Password, Name: req.Body.Name})
	switch {
	case errors.Is(err, app.ErrAccountAlreadyExists):
		return publicapi.Register409JSONResponse(errorBody("account_exists", err)), nil
	case errors.Is(err, domain.ErrWeakPassword), errors.Is(err, app.ErrInvalidAccount):
		return publicapi.Register400JSONResponse{ErrorJSONResponse: errorBody("invalid_request", err)}, nil
	case err != nil:
		return nil, err
	}
	return publicapi.Register201JSONResponse{AccountID: openapitypes.UUID(accountID)}, nil
}

func (h handler) Login(ctx context.Context, req publicapi.LoginRequestObject) (publicapi.LoginResponseObject, error) {
	client := openapix.ClientFromContext(ctx)
	tokens, err := h.service.Login(ctx, app.LoginInput{
		Email:     req.Body.Email,
		Password:  req.Body.Password,
		UserAgent: client.UserAgent,
		IPAddress: client.IPAddress,
	})
	if errors.Is(err, domain.ErrInvalidCredentials) {
		return publicapi.Login401JSONResponse{ErrorJSONResponse: errorBody("invalid_credentials", err)}, nil
	}
	if err != nil {
		return nil, err
	}
	return publicapi.Login200JSONResponse{TokensJSONResponse: publicapi.TokensJSONResponse{
		AccessToken:           tokens.AccessToken,
		AccessTokenExpiresAt:  tokens.AccessTokenExpiresAt,
		RefreshToken:          tokens.RefreshToken,
		RefreshTokenExpiresAt: tokens.RefreshTokenExpiresAt,
		AccountID:             openapitypes.UUID(tokens.AccountID),
		SessionID:             openapitypes.UUID(tokens.SessionID),
	}}, nil
}

func (h handler) Logout(ctx context.Context, req publicapi.LogoutRequestObject) (publicapi.LogoutResponseObject, error) {
	err := h.service.Logout(ctx, app.LogoutInput{
		AccountID: uuid.UUID(req.Params.XAccountID),
		SessionID: uuid.UUID(req.Params.XSessionID),
	})
	if errors.Is(err, app.ErrSessionNotFound) {
		return publicapi.Logout401JSONResponse{ErrorJSONResponse: errorBody("no_session", err)}, nil
	}
	if err != nil {
		return nil, err
	}
	return publicapi.Logout204Response{}, nil
}

func (h handler) ChangePassword(ctx context.Context, req publicapi.ChangePasswordRequestObject) (publicapi.ChangePasswordResponseObject, error) {
	err := h.service.ChangePassword(ctx, app.ChangePasswordInput{
		AccountID:   uuid.UUID(req.Params.XAccountID),
		SessionID:   uuid.UUID(req.Params.XSessionID),
		OldPassword: req.Body.OldPassword,
		NewPassword: req.Body.NewPassword,
	})
	switch {
	case errors.Is(err, app.ErrSessionNotFound):
		return publicapi.ChangePassword401JSONResponse(errorBody("no_session", err)), nil
	case errors.Is(err, domain.ErrInvalidCredentials):
		return publicapi.ChangePassword401JSONResponse(errorBody("invalid_credentials", err)), nil
	case errors.Is(err, domain.ErrWeakPassword):
		return publicapi.ChangePassword400JSONResponse{ErrorJSONResponse: errorBody("weak_password", err)}, nil
	case err != nil:
		return nil, err
	}
	return publicapi.ChangePassword204Response{}, nil
}

func errorBody(code string, err error) publicapi.ErrorJSONResponse {
	return publicapi.ErrorJSONResponse{Code: code, Message: errors.Cause(err).Error()}
}
