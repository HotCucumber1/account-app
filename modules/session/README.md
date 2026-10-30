# Session

Сессии: создание с refresh-токеном (хранится SHA-256) и признаком нового устройства, быстрая
проверка «сессия жива?», отзыв одной сессии или всех сессий учётки кроме текущей.

## API

- gRPC `sessionInternalAPI.SessionService`: CreateSession, GetSession, ValidateSession, KillSession,
  KillAccountSessions.
- События: `SessionCreated`, `SessionClosed`.

## Настройки

| Переменная | По умолчанию |
| --- | --- |
| `SESSION_POLICY_TTL` | `720h` |

## Общие настройки

Префикс переменных — `SESSION_`. Подключение к MySQL и адреса серверов совпадают с шаблоном модуля:
`SESSION_MYSQL_*`, `SESSION_HTTP_ADDRESS` (`:8082`), `SESSION_GRPC_ADDRESS` (`:8081`). Публикация событий: `SESSION_AMQP_*`, `SESSION_RELAY_*`.

## Команды

```sh
mise run //modules/session:all
mise run //modules/session:integration-tests
mise run //modules/session:generate-migration "Описание"
```

Общая картина взаимодействия сервисов: [docs/core_services.md](../../docs/core_services.md).
