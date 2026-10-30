# Frontend

Личный кабинет на React (Vite + TypeScript). Все запросы идут через gateway:
`/auth/api/...` — регистрация, вход, выход; `/account/api/...` — профиль.

## Страницы

- `/register` — саморегистрация (`POST /auth/api/public/v1/auth/register`).
- `/login` — вход; access-токен хранится в `localStorage` и передаётся как `Authorization: Bearer`.
- `/` — личный кабинет: просмотр профиля, изменение имени (`PATCH /account/api/public/v1/accounts/me`), выход.
- `/product` — продукт: GIF из `public/product.gif`.

`/` и `/product` защищены `ProtectedRoute`: без токена — переход на `/login`; токен проверяется
запросом профиля, ответ 401 сбрасывает его.

## Запуск

```sh
mise run local-up        # из корня репозитория: модули и gateway на localhost:80
cd frontend
npm ci
npm run dev              # http://localhost:5173
```

Dev- и preview-сервер Vite проксируют `/auth/api` и `/account/api` на gateway; адрес задаётся
`GATEWAY_URL` (по умолчанию `http://localhost:80`).

```sh
npm run build            # проверка типов и сборка в dist/
npm run preview          # http://localhost:4173
```
