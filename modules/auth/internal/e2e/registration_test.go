package integration_tests

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	accountinternalapi "account/api/server/internalapi"
	"auth/api/server/publicapi"
)

const (
	weakPassword = "short"
	userAgent    = "Firefox"
)

var tooLongName = strings.Repeat("x", 256)

func registerWithResponse(t *testing.T, body publicapi.RegisterJSONRequestBody) *publicapi.RegisterResponse {
	t.Helper()
	response, err := publicClient(t, userAgent).RegisterWithResponse(t.Context(), body)
	require.NoError(t, err)
	return response
}

func requireAccountNotFound(t *testing.T, email string) {
	t.Helper()
	_, err := accountClient(t).FindAccountByEmail(t.Context(), &accountinternalapi.FindAccountByEmailRequest{Email: email})
	require.Equal(t, codes.NotFound, status.Code(err))
}

func TestRegistrationCreatesAccount(t *testing.T) {
	// Arrange
	email := uniqueEmail()

	// Act
	response := registerWithResponse(t, publicapi.RegisterJSONRequestBody{
		Email:    email,
		Password: password,
		Name:     "Test User",
	})

	// Assert
	require.Equal(t, http.StatusCreated, response.StatusCode(), string(response.Body))
	account, err := accountClient(t).GetAccountInfo(
		t.Context(),
		&accountinternalapi.GetAccountInfoRequest{
			AccountID: response.JSON201.AccountID.String(),
		},
	)
	require.NoError(t, err)
	require.Equal(t, email, account.Email)
	require.Equal(t, "Test User", account.Name)
}

func TestRegistrationCreatesCredentials(t *testing.T) {
	// Arrange
	email := uniqueEmail()

	// Act
	accountID := register(t, email)

	// Assert
	tokens := requireLogin(t, email, password, userAgent)
	require.Equal(t, accountID, tokens.AccountID.String())
}

func TestRegistrationWithDuplicateEmail(t *testing.T) {
	// Arrange
	email := uniqueEmail()
	accountID := register(t, email)

	// Act
	response := registerWithResponse(t, publicapi.RegisterJSONRequestBody{Email: email, Password: newPassword, Name: "Twin"})

	// Assert
	require.Equal(t, http.StatusConflict, response.StatusCode(), string(response.Body))
	account, err := accountClient(t).GetAccountInfo(t.Context(), &accountinternalapi.GetAccountInfoRequest{AccountID: accountID})
	require.NoError(t, err)
	require.Equal(t, "Test User", account.Name)
	require.Equal(t, http.StatusUnauthorized, login(t, email, newPassword, userAgent).StatusCode())
	require.Equal(t, accountID, requireLogin(t, email, password, userAgent).AccountID.String())
}

func TestRegistrationWithWeakPassword(t *testing.T) {
	// Arrange
	email := uniqueEmail()

	// Act
	response := registerWithResponse(t, publicapi.RegisterJSONRequestBody{Email: email, Password: weakPassword, Name: "Weak"})

	// Assert
	require.Equal(t, http.StatusBadRequest, response.StatusCode(), string(response.Body))
	requireAccountNotFound(t, email)
}

func TestRegistrationWithInvalidEmail(t *testing.T) {
	// Act
	response := registerWithResponse(t, publicapi.RegisterJSONRequestBody{Email: "broken", Password: password, Name: "Broken"})

	// Assert
	require.Equal(t, http.StatusBadRequest, response.StatusCode(), string(response.Body))
}

func TestRegistrationRejectedByAccountIsRolledBack(t *testing.T) {
	// Arrange
	email := uniqueEmail()

	// Act
	response := registerWithResponse(t, publicapi.RegisterJSONRequestBody{Email: email, Password: password, Name: tooLongName})

	// Assert
	require.Equal(t, http.StatusBadRequest, response.StatusCode(), string(response.Body))
	requireAccountNotFound(t, email)
	require.Equal(t, http.StatusUnauthorized, login(t, email, password, userAgent).StatusCode())
}

func TestRegistrationAfterRollbackSucceeds(t *testing.T) {
	// Arrange
	email := uniqueEmail()
	rejected := registerWithResponse(t, publicapi.RegisterJSONRequestBody{Email: email, Password: password, Name: tooLongName})
	require.Equal(t, http.StatusBadRequest, rejected.StatusCode(), string(rejected.Body))

	// Act
	accountID := register(t, email)

	// Assert
	require.Equal(t, accountID, requireLogin(t, email, password, userAgent).AccountID.String())
}
