package gateway

import (
	"context"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	api "github.com/distributed-programming-2026/lib/gateway/interceptor"

	"auth/internal/auth/app"
	appservice "auth/internal/auth/app/service"
)

const (
	headerAuthorization = "Authorization"
	headerAccountID     = "X-Account-ID"
	headerSessionID     = "X-Session-ID"
	bearerPrefix        = "bearer "
)

func NewHandler(service *appservice.AuthService) api.InterceptorServer {
	return &handler{
		service: service,
	}
}

type handler struct {
	service *appservice.AuthService
	api.UnimplementedInterceptorServer
}

func (h *handler) Intercept(ctx context.Context, req *api.InterceptRequest) (*api.InterceptResponse, error) {
	if req.GetRequestHeaders() == nil {
		return nil, status.Error(codes.InvalidArgument, "request headers phase required")
	}
	token, ok := bearerToken(req.GetRequestHeaders().GetHeaders())
	if !ok {
		return unauthorized(), nil
	}
	claims, err := h.service.AuthenticateRequest(ctx, token)
	if errors.Is(err, app.ErrInvalidToken) || errors.Is(err, app.ErrSessionNotFound) {
		return unauthorized(), nil
	}
	if err != nil {
		return nil, err
	}
	return &api.InterceptResponse{Result: &api.InterceptResponse_ContinueResponse{
		ContinueResponse: &api.ContinueResponse{HeaderMutation: &api.HeaderMutation{
			Remove: []string{headerAuthorization},
			Set: []*api.HeaderValues{
				{Key: headerAccountID, Values: []string{claims.AccountID.String()}},
				{Key: headerSessionID, Values: []string{claims.SessionID.String()}},
			},
		}},
	}}, nil
}

// TODO давай в куку
func bearerToken(headers []*api.Header) (string, bool) {
	for _, header := range headers {
		if !strings.EqualFold(header.GetKey(), headerAuthorization) {
			continue
		}
		value := header.GetValue()
		if len(value) <= len(bearerPrefix) || !strings.EqualFold(value[:len(bearerPrefix)], bearerPrefix) {
			return "", false
		}
		return strings.TrimSpace(value[len(bearerPrefix):]), true
	}
	return "", false
}

func unauthorized() *api.InterceptResponse {
	return &api.InterceptResponse{Result: &api.InterceptResponse_ImmediateResponse{
		ImmediateResponse: &api.ImmediateResponse{
			StatusCode: http.StatusUnauthorized,
			Headers: []*api.Header{
				{
					Key:   "Content-Type",
					Value: "application/json",
				},
			},
			Body: []byte(`{"code":"unauthorized","message":"valid access token required"}`),
		},
	}}
}
