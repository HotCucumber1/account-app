package integration_tests

import (
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/distributed-programming-2026/lib/common/testenv"

	sessionevents "session/api/events"
	internalapi "session/api/server/internalapi"
)

func TestCreateAndValidateSession(t *testing.T) {
	// Arrange
	accountID := newAccountID()

	// Act
	first := createSession(t, accountID, "Firefox")
	// Assert
	require.True(t, first.NewDevice)

	// Assert
	_, created := testenv.RequireEvent(t, env.App("session"), func(e *sessionevents.SessionCreated) bool {
		return e.SessionID == first.SessionID
	})
	require.Equal(t, accountID, created.AccountID)
	require.True(t, created.NewDevice)

	// Act & Assert
	require.False(t, createSession(t, accountID, "Firefox").NewDevice)
	require.True(t, createSession(t, accountID, "Chrome").NewDevice)

	// Act
	validation := validate(t, first.SessionID)
	// Assert
	require.True(t, validation.Valid)
	require.Equal(t, accountID, validation.AccountID)
	require.False(t, validate(t, uuid.New().String()).Valid)

	// Act
	session, err := sessionClient(t).GetSession(t.Context(), &internalapi.GetSessionRequest{SessionID: first.SessionID})
	// Assert
	require.NoError(t, err)
	require.Equal(t, "Firefox", session.UserAgent)
	require.Equal(t, "10.0.0.1", session.IPAddress)
	require.Equal(t, "active", session.Status)

	// Act
	_, err = sessionClient(t).GetSession(t.Context(), &internalapi.GetSessionRequest{SessionID: uuid.New().String()})
	// Assert
	require.Equal(t, codes.NotFound, status.Code(err))

	// Act
	_, err = sessionClient(t).CreateSession(t.Context(), &internalapi.CreateSessionRequest{AccountID: accountID, IPAddress: "bad"})
	// Assert
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestKillSessions(t *testing.T) {
	// Arrange
	accountID := newAccountID()
	current := createSession(t, accountID, "Firefox")
	other := createSession(t, accountID, "Chrome")
	third := createSession(t, accountID, "Edge")
	client := sessionClient(t)

	// Act
	_, err := client.KillSession(t.Context(), &internalapi.KillSessionRequest{SessionID: third.SessionID, Reason: "logout"})
	// Assert
	require.NoError(t, err)
	require.False(t, validate(t, third.SessionID).Valid)
	_, closed := testenv.RequireEvent(t, env.App("session"), func(e *sessionevents.SessionClosed) bool {
		return e.SessionID == third.SessionID
	})
	require.Equal(t, "logout", closed.Reason)

	// Act
	killed, err := client.KillAccountSessions(t.Context(), &internalapi.KillAccountSessionsRequest{
		AccountID:       accountID,
		ExceptSessionID: current.SessionID,
		Reason:          "password_changed",
	})
	// Assert
	require.NoError(t, err)
	require.Equal(t, int64(1), killed.Killed)
	require.True(t, validate(t, current.SessionID).Valid)
	require.False(t, validate(t, other.SessionID).Valid)

	// Act
	session, err := client.GetSession(t.Context(), &internalapi.GetSessionRequest{SessionID: other.SessionID})
	// Assert
	require.NoError(t, err)
	require.Equal(t, "revoked", session.Status)
	require.Equal(t, "password_changed", session.RevokeReason)

	// Act
	_, err = client.KillSession(t.Context(), &internalapi.KillSessionRequest{SessionID: current.SessionID, Reason: "unknown"})
	// Assert
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestExpiredSessionIsInvalid(t *testing.T) {
	// Arrange
	created := createSession(t, newAccountID(), "Opera")

	// Act
	_, err := env.App("session").DB.ExecContext(t.Context(), "test.expire",
		"UPDATE session SET expires_at = UTC_TIMESTAMP(6) - INTERVAL 1 SECOND WHERE id = UUID_TO_BIN(?)", created.SessionID)
	require.NoError(t, err)

	// Assert
	require.False(t, validate(t, created.SessionID).Valid)

	// Act
	session, err := sessionClient(t).GetSession(t.Context(), &internalapi.GetSessionRequest{SessionID: created.SessionID})
	// Assert
	require.NoError(t, err)
	require.Equal(t, "expired", session.Status)
}
