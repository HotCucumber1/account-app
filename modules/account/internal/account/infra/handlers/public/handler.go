package public

import (
	"context"
	"uuid"

	openapitypes "github.com/oapi-codegen/runtime/types"
	"github.com/pkg/errors"

	"account/api/server/publicapi"
	"account/internal/account/app"
	appquery "account/internal/account/app/query"
	appservice "account/internal/account/app/service"
	"account/internal/account/domain"
)

func NewHandler(
	service appservice.AccountService,
	queryService appquery.AccountQueryService,
) publicapi.StrictServerInterface {
	return &handler{
		service:      service,
		queryService: queryService,
	}
}

type handler struct {
	service      appservice.AccountService
	queryService appquery.AccountQueryService
}

func (h *handler) GetMyAccount(ctx context.Context, req publicapi.GetMyAccountRequestObject) (publicapi.GetMyAccountResponseObject, error) {
	account, err := h.queryService.GetAccount(ctx, uuid.UUID(req.Params.XAccountID))
	if errors.Is(err, domain.ErrAccountNotFound) {
		return publicapi.GetMyAccount404Response{}, nil
	}
	if err != nil {
		return nil, err
	}
	return publicapi.GetMyAccount200JSONResponse{AccountJSONResponse: toAccountResponse(account)}, nil
}

func (h *handler) UpdateMyAccount(ctx context.Context, req publicapi.UpdateMyAccountRequestObject) (publicapi.UpdateMyAccountResponseObject, error) {
	account, err := h.service.UpdateAccount(ctx, app.RenameAccountInput{
		AccountID: uuid.UUID(req.Params.XAccountID),
		Name:      req.Body.Name,
	})
	switch {
	case errors.Is(err, domain.ErrAccountNotFound):
		return publicapi.UpdateMyAccount404Response{}, nil
	case errors.Is(err, domain.ErrInvalidAccount):
		return publicapi.UpdateMyAccount400JSONResponse{ErrorJSONResponse: publicapi.ErrorJSONResponse{
			Code:    "invalid_request",
			Message: errors.Cause(err).Error(),
		}}, nil
	case err != nil:
		return nil, err
	}
	return publicapi.UpdateMyAccount200JSONResponse{AccountJSONResponse: toAccountResponse(account)}, nil
}

func toAccountResponse(account app.AccountData) publicapi.AccountJSONResponse {
	return publicapi.AccountJSONResponse{
		Id:        openapitypes.UUID(account.ID),
		Email:     account.Email,
		Name:      account.Name,
		CreatedAt: account.CreatedAt,
	}
}
