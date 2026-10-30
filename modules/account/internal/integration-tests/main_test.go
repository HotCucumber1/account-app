package integration_tests

import (
	"os"
	"testing"

	"github.com/distributed-programming-2026/lib/common/testenv"
)

func TestMain(m *testing.M) {
	os.Exit(testenv.Main(m, &env, testenv.Module{
		Name:     "account",
		Commands: []string{testenv.CommandService},
	}))
}
