# Account

Учётки пользователей: создание, чтение и изменение имени. Не хранит пароли, токены, сессии и роли.

## API

- gRPC `accountInternalAPI.AccountService`: CreateAccount, GetAccountInfo, FindAccountByEmail.
- HTTP (за gateway, заголовок `X-Account-ID`): `GET /api/public/v1/accounts/me`,
  `PATCH /api/public/v1/accounts/me` (`{"name": "..."}`; 400 при пустом или длинном имени).
  Учётка определяется только по `X-Account-ID`, который выставляет интерцептор Auth.
- Gateway: `config/gateway/routes.yaml` — `/account/api/public/v1/accounts/me` (GET, PATCH)
  через интерцептор `auth-interceptor`.
- События: `AccountCreated`.

## Общие настройки

Префикс переменных — `ACCOUNT_`. Подключение к MySQL и адреса серверов совпадают с шаблоном модуля:
`ACCOUNT_MYSQL_*`, `ACCOUNT_HTTP_ADDRESS` (`:8082`), `ACCOUNT_GRPC_ADDRESS` (`:8081`). Публикация событий: `ACCOUNT_AMQP_*`, `ACCOUNT_RELAY_*`.

## Команды

```sh
mise run //modules/account:all
mise run //modules/account:integration-tests
mise run //modules/account:generate-migration "Описание"
```

Общая картина взаимодействия сервисов: [docs/core_services.md](../../docs/core_services.md).
