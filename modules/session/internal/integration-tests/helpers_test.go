package integration_tests

import (
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/distributed-programming-2026/lib/common/grpcx"

	internalapi "session/api/server/internalapi"
)

func sessionClient(t *testing.T) internalapi.SessionServiceClient {
	t.Helper()
	conn, err := grpcx.Dial(env.App("session").GRPCAddress)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return internalapi.NewSessionServiceClient(conn)
}

func createSession(t *testing.T, accountID, userAgent string) *internalapi.CreateSessionResponse {
	t.Helper()
	created, err := sessionClient(t).CreateSession(t.Context(), &internalapi.CreateSessionRequest{
		AccountID: accountID,
		UserAgent: userAgent,
		IPAddress: "10.0.0.1",
	}, grpc.WaitForReady(true))
	require.NoError(t, err)
	return created
}

func validate(t *testing.T, sessionID string) *internalapi.ValidateSessionResponse {
	t.Helper()
	validation, err := sessionClient(t).ValidateSession(t.Context(), &internalapi.ValidateSessionRequest{SessionID: sessionID})
	require.NoError(t, err)
	return validation
}

func newAccountID() string {
	return uuid.New().String()
}
