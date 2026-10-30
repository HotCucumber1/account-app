package integration_tests

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/distributed-programming-2026/lib/common/testenv"

	accountevents "account/api/events"
	accountinternalapi "account/api/server/internalapi"
	authevents "auth/api/events"
	authinternalapi "auth/api/server/internalapi"
	"auth/api/server/publicapi"
	sessionevents "session/api/events"
)

func TestRegistration(t *testing.T) {
	// Arrange
	email := uniqueEmail()

	// Act
	accountID := register(t, email)

	// Assert
	account, err := accountClient(t).GetAccountInfo(t.Context(), &accountinternalapi.GetAccountInfoRequest{AccountID: accountID})
	require.NoError(t, err)
	require.Equal(t, email, account.Email)
	testenv.RequireEvent(t, env.App("account"), func(e *accountevents.AccountCreated) bool { return e.AccountID == accountID })
	testenv.RequireEvent(t, env.App("auth"), func(e *authevents.UserRegistered) bool { return e.AccountID == accountID })

	// Arrange
	client := publicClient(t, "Firefox")

	// Act
	duplicate, err := client.RegisterWithResponse(t.Context(), publicapi.RegisterJSONRequestBody{Email: email, Password: password, Name: "Twin"})
	// Assert
	require.NoError(t, err)
	require.Equal(t, http.StatusConflict, duplicate.StatusCode())

	// Arrange
	weakEmail := uniqueEmail()

	// Act
	weak, err := client.RegisterWithResponse(t.Context(), publicapi.RegisterJSONRequestBody{Email: weakEmail, Password: "short", Name: "Weak"})
	// Assert
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, weak.StatusCode())
	_, err = accountClient(t).FindAccountByEmail(t.Context(), &accountinternalapi.FindAccountByEmailRequest{Email: weakEmail})
	require.Equal(t, codes.NotFound, status.Code(err))

	// Act
	invalid, err := client.RegisterWithResponse(t.Context(), publicapi.RegisterJSONRequestBody{Email: "broken", Password: password, Name: "Broken"})
	// Assert
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, invalid.StatusCode())
}

func TestLogin(t *testing.T) {
	// Arrange
	email := uniqueEmail()
	accountID := register(t, email)

	// Act
	tokens := requireLogin(t, strings.ToUpper(email), password, "Firefox")
	// Assert
	require.Equal(t, accountID, tokens.AccountID.String())
	require.True(t, sessionValid(t, tokens.SessionID))

	// Act
	claims, err := authClient(t).ValidateToken(t.Context(), &authinternalapi.ValidateTokenRequest{AccessToken: tokens.AccessToken})
	// Assert
	require.NoError(t, err)
	require.Equal(t, accountID, claims.AccountID)
	require.Equal(t, tokens.SessionID.String(), claims.SessionID)

	// Assert
	_, loggedIn := testenv.RequireEvent(t, env.App("auth"), func(e *authevents.UserLoggedIn) bool {
		return e.SessionID == tokens.SessionID.String()
	})
	require.True(t, loggedIn.NewDevice)
	require.Equal(t, clientIP, loggedIn.IPAddress)
	require.Equal(t, "Firefox", loggedIn.UserAgent)
	testenv.RequireEvent(t, env.App("session"), func(e *sessionevents.SessionCreated) bool {
		return e.SessionID == tokens.SessionID.String()
	})

	// Act
	again := requireLogin(t, email, password, "Firefox")
	// Assert
	_, loggedIn = testenv.RequireEvent(t, env.App("auth"), func(e *authevents.UserLoggedIn) bool {
		return e.SessionID == again.SessionID.String()
	})
	require.False(t, loggedIn.NewDevice)

	// Act
	_, err = authClient(t).ValidateToken(t.Context(), &authinternalapi.ValidateTokenRequest{AccessToken: tokens.AccessToken + "x"})
	// Assert
	require.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestLoginWithWrongPassword(t *testing.T) {
	// Arrange
	email := uniqueEmail()
	accountID := register(t, email)

	// Act
	require.Equal(t, http.StatusUnauthorized, login(t, email, "wrong-password", "Firefox").StatusCode())
	// Assert
	_, failed := testenv.RequireEvent(t, env.App("auth"), func(e *authevents.UserLoginFailed) bool {
		return e.AccountID == accountID
	})
	require.Equal(t, "invalid_password", failed.Reason)
	require.Equal(t, clientIP, failed.IPAddress)

	// Arrange
	unknown := uniqueEmail()

	// Act
	require.Equal(t, http.StatusUnauthorized, login(t, unknown, password, "Firefox").StatusCode())
	// Assert
	_, failed = testenv.RequireEvent(t, env.App("auth"), func(e *authevents.UserLoginFailed) bool {
		return e.Email == unknown
	})
	require.Equal(t, "unknown_account", failed.Reason)
	require.Empty(t, failed.AccountID)
}

func TestLogout(t *testing.T) {
	// Arrange
	email := uniqueEmail()
	register(t, email)
	tokens := requireLogin(t, email, password, "Firefox")
	client := publicClient(t, "Firefox")
	params := &publicapi.LogoutParams{XAccountID: tokens.AccountID, XSessionID: tokens.SessionID}

	// Act
	logout, err := client.LogoutWithResponse(t.Context(), params)
	// Assert
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, logout.StatusCode())
	require.False(t, sessionValid(t, tokens.SessionID))
	testenv.RequireEvent(t, env.App("auth"), func(e *authevents.UserLoggedOut) bool { return e.SessionID == tokens.SessionID.String() })
	_, closed := testenv.RequireEvent(t, env.App("session"), func(e *sessionevents.SessionClosed) bool {
		return e.SessionID == tokens.SessionID.String()
	})
	require.Equal(t, "logout", closed.Reason)

	// Act
	logout, err = client.LogoutWithResponse(t.Context(), params)
	// Assert
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, logout.StatusCode())
}

func TestChangePassword(t *testing.T) {
	// Arrange
	email := uniqueEmail()
	register(t, email)
	current := requireLogin(t, email, password, "Firefox")
	other := requireLogin(t, email, password, "Chrome")
	client := publicClient(t, "Firefox")
	params := &publicapi.ChangePasswordParams{XAccountID: current.AccountID, XSessionID: current.SessionID}

	// Act
	wrong, err := client.ChangePasswordWithResponse(t.Context(), params, publicapi.ChangePasswordJSONRequestBody{OldPassword: "wrong-password", NewPassword: newPassword})
	// Assert
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, wrong.StatusCode())

	// Act
	weak, err := client.ChangePasswordWithResponse(t.Context(), params, publicapi.ChangePasswordJSONRequestBody{OldPassword: password, NewPassword: "short"})
	// Assert
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, weak.StatusCode())
	require.True(t, sessionValid(t, other.SessionID))

	// Act
	changed, err := client.ChangePasswordWithResponse(t.Context(), params, publicapi.ChangePasswordJSONRequestBody{OldPassword: password, NewPassword: newPassword})
	// Assert
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, changed.StatusCode())
	require.True(t, sessionValid(t, current.SessionID))
	require.False(t, sessionValid(t, other.SessionID))
	testenv.RequireEvent(t, env.App("auth"), func(e *authevents.PasswordChanged) bool { return e.AccountID == current.AccountID.String() })
	_, closed := testenv.RequireEvent(t, env.App("session"), func(e *sessionevents.SessionClosed) bool {
		return e.SessionID == other.SessionID.String()
	})
	require.Equal(t, "password_changed", closed.Reason)

	// Act & Assert
	require.Equal(t, http.StatusUnauthorized, login(t, email, password, "Firefox").StatusCode())
	requireLogin(t, email, newPassword, "Firefox")
}
