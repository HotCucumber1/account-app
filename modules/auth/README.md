# Auth

Регистрация, вход, выход, проверка access-токена и смена пароля. Хранит хэши паролей (bcrypt)
и ключи подписи. Access-токен — JWT EdDSA с `sub` = учётка и `sid` = сессия; ключ создаётся при
первом запуске под именованной блокировкой. Refresh-токен выпускает Session.

## API

- gRPC `authInternalAPI.AuthService`: Register, Authenticate, RevokeToken, ValidateToken, ChangePassword.
- HTTP: `POST /api/public/v1/auth/register`, `/login`, `/logout`, `/password/change`.
  Logout и смена пароля требуют `X-Account-ID` и `X-Session-ID` от gateway.
- gRPC `interceptorgrpc.Interceptor` для gateway: проверяет `Authorization: Bearer <access-токен>`
  и живость сессии, заменяет `X-Account-ID`/`X-Session-ID` значениями из токена и удаляет
  `Authorization`; без валидного токена возвращает 401.
- Gateway: `config/gateway/interceptor.yaml` объявляет `auth-interceptor`, `routes.yaml` —
  `/auth/api/public/v1/auth/register|login` без проверки и `logout|password/change` через интерцептор.
- События: `UserRegistered`, `UserLoggedIn`, `UserLoginFailed`, `UserLoggedOut`, `PasswordChanged`.

## Настройки

| Переменная                                     | По умолчанию           |
|------------------------------------------------|------------------------|
| `AUTH_ACCOUNT_GRPC_ADDRESS`                    | `account-service:8081` |
| `AUTH_SESSION_GRPC_ADDRESS`                    | `session-service:8081` |
| `AUTH_ACCOUNT_TIMEOUT`, `AUTH_SESSION_TIMEOUT` | `5s`                   |
| `AUTH_TOKEN_ACCESS_TTL`                        | `15m`                  |
| `AUTH_TOKEN_KEY_LOCK_TIMEOUT`                  | `10s`                  |
| `AUTH_PASSWORD_BCRYPT_COST`                    | `10`                   |

## Общие настройки

Префикс переменных — `AUTH_`. Подключение к MySQL и адреса серверов совпадают с шаблоном модуля:
`AUTH_MYSQL_*`, `AUTH_HTTP_ADDRESS` (`:8082`), `AUTH_GRPC_ADDRESS` (`:8081`). Публикация событий: `AUTH_AMQP_*`,
`AUTH_RELAY_*`.

## Команды

```sh
mise run //modules/auth:all
mise run //modules/auth:integration-tests
mise run //modules/auth:generate-migration "Описание"
```

Общая картина взаимодействия сервисов: [docs/core_services.md](../../docs/core_services.md).
