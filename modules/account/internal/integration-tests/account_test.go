package integration_tests

import (
	"net/http"
	"strings"
	"testing"
	"uuid"

	openapitypes "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/distributed-programming-2026/lib/common/testenv"

	accountevents "account/api/events"
	internalapi "account/api/server/internalapi"
	"account/api/server/publicapi"
)

func TestCreateAccount(t *testing.T) {
	// Arrange
	email := uniqueEmail()

	// Act
	id := createAccount(t, email)

	// Assert
	_, created := testenv.RequireEvent(t, env.App("account"), func(e *accountevents.AccountCreated) bool {
		return e.AccountID == id
	})
	require.Equal(t, email, created.Email)
	require.Equal(t, "Test User", created.Name)

	// Arrange
	client := accountClient(t)

	// Act
	_, err := client.CreateAccount(t.Context(), &internalapi.CreateAccountRequest{Email: strings.ToUpper(email), Name: "Twin"})
	// Assert
	require.Equal(t, codes.AlreadyExists, status.Code(err))

	// Act
	_, err = client.CreateAccount(t.Context(), &internalapi.CreateAccountRequest{Email: "broken", Name: "Twin"})
	// Assert
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	// Act
	_, err = client.CreateAccount(t.Context(), &internalapi.CreateAccountRequest{Email: uniqueEmail(), Name: " "})
	// Assert
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestGetAccount(t *testing.T) {
	// Arrange
	email := uniqueEmail()
	id := createAccount(t, email)
	client := accountClient(t)

	// Act
	info, err := client.GetAccountInfo(t.Context(), &internalapi.GetAccountInfoRequest{AccountID: id})
	// Assert
	require.NoError(t, err)
	require.Equal(t, email, info.Email)
	require.Equal(t, "Test User", info.Name)

	// Act
	found, err := client.FindAccountByEmail(t.Context(), &internalapi.FindAccountByEmailRequest{Email: " " + strings.ToUpper(email)})
	// Assert
	require.NoError(t, err)
	require.Equal(t, id, found.ID)

	// Act
	_, err = client.FindAccountByEmail(t.Context(), &internalapi.FindAccountByEmailRequest{Email: uniqueEmail()})
	// Assert
	require.Equal(t, codes.NotFound, status.Code(err))

	// Act
	_, err = client.GetAccountInfo(t.Context(), &internalapi.GetAccountInfoRequest{AccountID: "bad"})
	// Assert
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	// Act
	me, err := publicClient(t).GetMyAccountWithResponse(t.Context(), &publicapi.GetMyAccountParams{
		XAccountID: openapitypes.UUID(uuid.MustParse(id)),
	})
	// Assert
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, me.StatusCode())
	require.Equal(t, email, me.JSON200.Email)

	// Act
	missing, err := publicClient(t).GetMyAccountWithResponse(t.Context(), &publicapi.GetMyAccountParams{
		XAccountID: openapitypes.UUID(uuid.New()),
	})
	// Assert
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, missing.StatusCode())
}

func TestUpdateMyAccount(t *testing.T) {
	// Arrange
	email := uniqueEmail()
	id := createAccount(t, email)
	otherID := createAccount(t, uniqueEmail())
	client := publicClient(t)
	params := &publicapi.UpdateMyAccountParams{XAccountID: openapitypes.UUID(uuid.MustParse(id))}

	// Act
	updated, err := client.UpdateMyAccountWithResponse(t.Context(), params, publicapi.UpdateMyAccountJSONRequestBody{Name: "  Renamed User "})

	// Assert
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, updated.StatusCode(), string(updated.Body))
	require.Equal(t, "Renamed User", updated.JSON200.Name)
	require.Equal(t, email, updated.JSON200.Email)
	require.Equal(t, id, updated.JSON200.Id.String())

	info, err := accountClient(t).GetAccountInfo(t.Context(), &internalapi.GetAccountInfoRequest{AccountID: id})
	require.NoError(t, err)
	require.Equal(t, "Renamed User", info.Name)

	other, err := accountClient(t).GetAccountInfo(t.Context(), &internalapi.GetAccountInfoRequest{AccountID: otherID})
	require.NoError(t, err)
	require.Equal(t, "Test User", other.Name)

	for _, name := range []string{"", "   ", strings.Repeat("я", 256)} {
		// Act
		invalid, err := client.UpdateMyAccountWithResponse(t.Context(), params, publicapi.UpdateMyAccountJSONRequestBody{Name: name})
		// Assert
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, invalid.StatusCode())
		require.Equal(t, "invalid_request", invalid.JSON400.Code)
	}
	info, err = accountClient(t).GetAccountInfo(t.Context(), &internalapi.GetAccountInfoRequest{AccountID: id})
	require.NoError(t, err)
	require.Equal(t, "Renamed User", info.Name)

	// Act
	missing, err := client.UpdateMyAccountWithResponse(t.Context(), &publicapi.UpdateMyAccountParams{
		XAccountID: openapitypes.UUID(uuid.New()),
	}, publicapi.UpdateMyAccountJSONRequestBody{Name: "Ghost"})
	// Assert
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, missing.StatusCode())

	// Act
	request, err := http.NewRequestWithContext(t.Context(), http.MethodPatch, "http://"+env.App("account").HTTPAddress+"/api/public/v1/accounts/me",
		strings.NewReader(`{"name":"Anonymous"}`),
	)
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	anonymous, err := http.DefaultClient.Do(request)
	// Assert
	require.NoError(t, err)
	require.NoError(t, anonymous.Body.Close())
	require.Equal(t, http.StatusBadRequest, anonymous.StatusCode)
}
