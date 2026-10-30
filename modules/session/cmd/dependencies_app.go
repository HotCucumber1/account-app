package main

import (
	"github.com/distributed-programming-2026/go-sdk/pkg/event"

	"github.com/distributed-programming-2026/lib/common/uowx"

	appquery "session/internal/session/app/query"
	appservice "session/internal/session/app/service"
	"session/internal/session/config"
	domainservice "session/internal/session/domain/service"
	infraevent "session/internal/session/infra/event"
	"session/internal/session/infra/mysql/query"
	"session/internal/session/infra/mysql/repository"
)

type services struct {
	service      *appservice.SessionService
	queryService appquery.SessionQueryService
}

func newServices(conf config.Env, db *database, dispatcher event.Dispatcher) *services {
	return &services{
		service: appservice.NewSessionService(
			uowx.NewUnitOfWork(db.unit),
			domainservice.NewSessionService(
				repository.NewSessionRepository(db.unit),
				repository.NewDeviceRepository(db.unit),
				infraevent.NewDispatcher(infraevent.NewDispatcherFactory(appID, dispatcher)),
				conf.Policy.TTL,
			),
		),
		queryService: query.NewSessionQueryService(db.unit),
	}
}
