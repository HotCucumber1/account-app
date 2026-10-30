package testenv

import (
	"context"
	"fmt"
	"os"
	"testing"
)

// Main is a TestMain helper: it starts the environment, assigns it to target and runs tests.
func Main(m *testing.M, target **Environment, modules ...Module) int {
	// Container log streaming is bound to this context, so it lives until tests finish;
	// every container has its own startup timeout.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	env, err := Start(ctx, modules...)
	if err != nil {
		fmt.Fprintln(os.Stderr, env.Logs())
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	*target = env

	code := m.Run()
	if code != 0 {
		fmt.Fprintln(os.Stderr, env.Logs())
	}
	err = env.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return code
}
