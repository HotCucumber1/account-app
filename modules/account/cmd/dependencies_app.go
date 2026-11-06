package main

import (
	"github.com/distributed-programming-2026/go-sdk/pkg/event"

	"github.com/distributed-programming-2026/lib/common/uowx"

	appquery "account/internal/account/app/query"
	appservice "account/internal/account/app/service"
	domainservice "account/internal/account/domain/service"
	infraevent "account/internal/account/infra/event"
	"account/internal/account/infra/mysql/query"
	"account/internal/account/infra/mysql/repository"
	"account/internal/account/infra/mysql/xa"
)

type services struct {
	service      *appservice.AccountService
	queryService appquery.AccountQueryService
}

func newServices(db *database, dispatcher event.Dispatcher) *services {
	return &services{
		service: appservice.NewAccountService(
			uowx.NewUnitOfWork(db.unit),
			xa.NewBranches(db.client, appID),
			domainservice.NewAccountService(repository.NewAccountRepository(xa.Unit(db.unit)), infraevent.NewDispatcher(infraevent.NewDispatcherFactory(appID, dispatcher))),
		),
		queryService: query.NewAccountQueryService(db.unit),
	}
}
