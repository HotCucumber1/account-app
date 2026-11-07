package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/distributed-programming-2026/go-sdk/pkg/logging"

	appservice "auth/internal/auth/app/service"
	"auth/internal/auth/config"
)

func runRegistrationRecovery(
	ctx context.Context,
	logger *slog.Logger,
	coordinator *appservice.RegistrationCoordinator,
	conf config.Registration,
) error {
	ticker := time.NewTicker(conf.RecoveryInterval)
	defer ticker.Stop()

	for {
		err := coordinator.Recover(ctx)
		if err != nil && ctx.Err() == nil {
			logger.Error("registration recovery failed; will retry", logging.Error(err))
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
