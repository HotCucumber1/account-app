package integration_tests

import (
	"os"
	"testing"

	"github.com/distributed-programming-2026/lib/common/testenv"
)

var env *testenv.Environment

func TestMain(m *testing.M) {
	service := []string{testenv.CommandService}
	os.Exit(testenv.Main(m, &env,
		testenv.Module{Name: "account", Commands: service},
		testenv.Module{Name: "session", Commands: service},
		testenv.Module{
			Name:     "auth",
			Commands: service,
			Env: map[string]string{
				"ACCOUNT_GRPC_ADDRESS": "account-service:8081",
				"SESSION_GRPC_ADDRESS": "session-service:8081",
				"PASSWORD_BCRYPT_COST": "4",
			},
		},
	))
}
