package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/distributed-programming-2026/go-sdk/pkg/event"

	"github.com/distributed-programming-2026/lib/common/grpcx"
	"github.com/distributed-programming-2026/lib/common/uowx"

	appservice "auth/internal/auth/app/service"
	"auth/internal/auth/config"
	domainservice "auth/internal/auth/domain/service"
	"auth/internal/auth/infra/account"
	infraevent "auth/internal/auth/infra/event"
	"auth/internal/auth/infra/mysql/repository"
	"auth/internal/auth/infra/password"
	"auth/internal/auth/infra/session"
	"auth/internal/auth/infra/token"

	"github.com/distributed-programming-2026/lib/common/xa"
)

type services struct {
	service      *appservice.AuthService
	registration *appservice.RegistrationCoordinator
}

func newServices(
	_ context.Context,
	logger *slog.Logger,
	conf config.Env,
	db *database,
	dispatcher event.Dispatcher,
) (*services, grpcx.Connections, error) {
	unitOfWork := uowx.NewUnitOfWork(db.unit)
	tokenIssuer := token.NewMockJWT(conf.Token.AccessTTL)
	credentialService := domainservice.NewAuthService(
		repository.NewCredentialRepository(xa.Unit(db.unit)),
		password.NewBcrypt(conf.Password.BcryptCost),
		infraevent.NewDispatcher(infraevent.NewDispatcherFactory(appID, dispatcher)),
	)

	var conns grpcx.Connections
	accountConn, err := conns.Dial(conf.Account.GRPCAddress)
	if err != nil {
		return nil, nil, err
	}
	sessionConn, err := conns.Dial(conf.Session.GRPCAddress)
	if err != nil {
		return nil, nil, errors.Join(err, conns.Close())
	}
	accountClient := account.NewClient(accountConn, conf.Account.Timeout)
	sessionClient := session.NewClient(sessionConn, conf.Session.Timeout)
	registration := appservice.NewRegistrationCoordinator(
		unitOfWork,
		repository.NewRegistrationRepository(db.unit),
		[]appservice.RegistrationParticipant{
			accountClient,
			appservice.NewCredentialParticipant(xa.NewBranches(db.client, appID), credentialService),
		},
		appservice.RegistrationConfig{
			Deadline:      conf.Registration.Deadline,
			RecoveryGrace: conf.Registration.RecoveryGrace,
			RecoveryBatch: conf.Registration.RecoveryBatch,
		},
		logger,
	)
	return &services{
		service: appservice.NewAuthService(
			unitOfWork,
			credentialService,
			accountClient,
			registration,
			sessionClient,
			sessionClient,
			tokenIssuer,
		),
		registration: registration,
	}, conns, nil
}
