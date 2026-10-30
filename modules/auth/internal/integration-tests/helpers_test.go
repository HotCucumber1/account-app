package integration_tests

import (
	"context"
	"net/http"
	"testing"
	"uuid"

	openapitypes "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/distributed-programming-2026/lib/common/grpcx"

	accountinternalapi "account/api/server/internalapi"
	authinternalapi "auth/api/server/internalapi"
	publicapi "auth/api/server/publicapi"
	sessioninternalapi "session/api/server/internalapi"
)

const (
	password    = "correct-horse"
	newPassword = "battery-staple"
	clientIP    = "192.168.10.20"
)

func dial(t *testing.T, module string) *grpc.ClientConn {
	t.Helper()
	conn, err := grpcx.Dial(env.App(module).GRPCAddress)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func authClient(t *testing.T) authinternalapi.AuthServiceClient {
	t.Helper()
	return authinternalapi.NewAuthServiceClient(dial(t, "auth"))
}

func accountClient(t *testing.T) accountinternalapi.AccountServiceClient {
	t.Helper()
	return accountinternalapi.NewAccountServiceClient(dial(t, "account"))
}

func sessionClient(t *testing.T) sessioninternalapi.SessionServiceClient {
	t.Helper()
	return sessioninternalapi.NewSessionServiceClient(dial(t, "session"))
}

func publicClient(t *testing.T, userAgent string) *publicapi.ClientWithResponses {
	t.Helper()
	client, err := publicapi.NewClientWithResponses("http://"+env.App("auth").HTTPAddress,
		publicapi.WithRequestEditorFn(func(_ context.Context, r *http.Request) error {
			r.Header.Set("User-Agent", userAgent)
			r.Header.Set("X-Real-IP", clientIP)
			return nil
		}))
	require.NoError(t, err)
	return client
}

func uniqueEmail() string {
	return "user-" + uuid.New().String() + "@example.com"
}

func register(t *testing.T, email string) string {
	t.Helper()
	response, err := publicClient(t, "Firefox").RegisterWithResponse(t.Context(), publicapi.RegisterJSONRequestBody{
		Email:    email,
		Password: password,
		Name:     "Test User",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, response.StatusCode(), string(response.Body))
	return response.JSON201.AccountID.String()
}

func login(t *testing.T, email, pass, userAgent string) *publicapi.LoginResponse {
	t.Helper()
	response, err := publicClient(t, userAgent).LoginWithResponse(t.Context(), publicapi.LoginJSONRequestBody{
		Email:    email,
		Password: pass,
	})
	require.NoError(t, err)
	return response
}

func requireLogin(t *testing.T, email, pass, userAgent string) *publicapi.Tokens {
	t.Helper()
	response := login(t, email, pass, userAgent)
	require.Equal(t, http.StatusOK, response.StatusCode(), string(response.Body))
	return response.JSON200
}

func sessionValid(t *testing.T, sessionID openapitypes.UUID) bool {
	t.Helper()
	validation, err := sessionClient(t).ValidateSession(t.Context(), &sessioninternalapi.ValidateSessionRequest{SessionID: sessionID.String()})
	require.NoError(t, err)
	return validation.Valid
}
