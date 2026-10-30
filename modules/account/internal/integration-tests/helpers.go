package integration_tests

import (
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/distributed-programming-2026/lib/common/grpcx"
	"github.com/distributed-programming-2026/lib/common/testenv"

	internalapi "account/api/server/internalapi"
	"account/api/server/publicapi"
)

var env *testenv.Environment

func accountClient(t *testing.T) internalapi.AccountServiceClient {
	t.Helper()
	conn, err := grpcx.Dial(env.App("account").GRPCAddress)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return internalapi.NewAccountServiceClient(conn)
}

func publicClient(t *testing.T) *publicapi.ClientWithResponses {
	t.Helper()
	client, err := publicapi.NewClientWithResponses("http://" + env.App("account").HTTPAddress)
	require.NoError(t, err)
	return client
}

func uniqueEmail() string {
	return "user-" + uuid.New().String() + "@example.com"
}

func createAccount(t *testing.T, email string) string {
	t.Helper()
	created, err := accountClient(t).CreateAccount(t.Context(), &internalapi.CreateAccountRequest{
		Email: email,
		Name:  "Test User",
	}, grpc.WaitForReady(true))
	require.NoError(t, err)
	return created.AccountID
}
