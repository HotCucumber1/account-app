# Лабораторная работа №2 — Решение

## Немножко напоминания владения

| Данные         | Владелец |
|----------------|----------|
| User           | Account  |
| Email, profile | Account  |
| Password Hash  | Auth     |
| Session        | Session  |
| Roles          | Access   |
| Audit Records  | Audit    |

## Граница Sync / Async

```
Client
 |
Gateway
 |
Auth
 |
Account
 |
Session

=================
|| ответ юзеру ||
=================

Broker
 |
Audit
```

## Сценарии и взаимодействие сервисов

### Регистрация (первая)

```mermaid
sequenceDiagram
    autonumber
    actor Client
    participant GW as Gateway
    participant Auth
    participant Auth_DB
    participant Account
    participant Account_DB
    participant Session
    participant Session_DB
    participant Broker
    participant Audit

    Client ->> GW: POST /register
    Note over GW: generate request_id
    GW ->> Auth: Register(email,password)
    Auth ->> Account: FindByEmail
    Account ->> Account_DB: find()
    Account_DB -->> Account: data
    Account -->> Auth: data
    
    alt has_user
        Auth -->> GW: USER_ALREADY_EXISTS
        GW -->> Client: 409 Conflict
    else no
        Auth ->> Account: CreateAccount()
        Account ->> Account_DB: store()
        Account_DB -->> Account: user_id
        Account -->> Auth: user_id
        Auth ->> Auth: hash password
        Auth ->> Auth: save credentials
        Auth ->> Auth_DB: store()
        Auth -->> GW: OK
        GW -->> Client: 201
        Note over Auth, Audit: Async Part (после создания)
        Auth -->> Broker: UserRegistered
        Broker -->> Audit: UserRegistered
        Audit ->> Audit: save audit

        Client ->> GW: /login
    end
```

## Login

[//]: # (TODO Чек сессиии, чтобы я не мог создать сессиию, если послал запрос логинации будучи уже залогированным)

```mermaid
sequenceDiagram
    autonumber
    actor Client
    participant GW as Gateway
    participant Auth
    participant Auth_DB
    participant Account
    participant Account_DB
    participant Session
    participant Session_DB
    participant Broker
    participant Audit

    Client ->> GW: POST /login
    Note over GW: request_id
    
    GW ->> Session: CheckSession()
    Session -->> GW: SessionData
    
    alt has_session
        GW -->> Client: 302 redirect
    else
        GW ->> Auth: Authenticate(email,password)
        Auth ->> Account: FindByEmail
        Account ->> Account_DB: find()
        Account_DB -->> Account: Data
        Account -->> Auth: user_id
        Auth ->> Auth_DB: find()
        Auth_DB -->> Auth: password_hash
        Auth ->> Auth: verify password hash
        Auth ->> Session: CreateSession(user_id)
        Session ->> Session_DB: store()
        Session_DB -->> Session: session_id
        Session -->> Auth: session_id
        Auth -->> GW: access_token, refresh_token
        GW -->> Client: 200 OK
        Note over Auth, Audit: Async Part (после логинки)
        Session -->> Broker: EnterAccepted
        Broker -->> Audit: EnterAccepted
        Audit ->> Audit: save record
    end
```

## Login с не тем паролем

```mermaid
sequenceDiagram
    actor Client
    participant GW as Gateway
    participant Auth
    participant Auth_DB
    participant Account
    participant Account_DB
    participant Broker
    participant Audit
    
    Client ->> GW: POST /login
    Note over GW: request_id
    GW ->> Auth: Login
    Auth ->> Account: FindByEmail
    Account ->> Account_DB: find()
    Account_DB -->> Account: user_id
    Account -->> Auth: user_id
    Auth ->> Auth: verify hash
    Auth ->> Auth_DB: find()
    Auth_DB -->> Auth: password_hash
    Auth -->> GW: INVALID
    GW -->> Client: INVALID
    Auth -->> Broker: UserLoginFailed
    Broker -->> Audit: UserLoginFailed
```

## Продукт юзает это так

[//]: # (TODO Ходить в Auth когда сессия закончила жизнь свою)
[//]: # (TODO Remember Me - это часть Auth, а не сессии. Это долгая кука)

```mermaid
sequenceDiagram
    actor Client
    participant GW as Gateway
    participant Session
    participant Access
    participant Product
    participant Broker
    participant Audit
    Client ->> GW: GET /product/resource
    GW ->> Session: CheckSession()
    Session -->> GW: SessionResponse()
    
    alt no_session
        GW -->> Client: 302 redirect
    else has
        GW -> Access: CheckAccess()
        Access -->> GW: has_access
        GW ->> Product: Request(user_id)
        Product -->> GW: response
        GW -->> Client: 200 OK
        Note over Session, Audit: Async Part (после логинки)
        Session -->> Broker: EnterAccepted
        Broker -->> Audit: EnterAccepted
        Audit ->> Audit: save record
    end
```

## Логаутим

```mermaid
sequenceDiagram
    autonumber
    actor Client
    participant GW as Gateway
    participant Auth
    participant Session
    participant Session_DB
    participant Broker
    participant Audit
    Client ->> GW: POST /logout
    GW ->> Session: CheckSession()
    Session -->> GW: SessionResponse()
    
    alt no_session
        GW -->> Client: 302 redirect
    else
        GW ->> Auth: Logout()
        Auth ->> Session: KillSession(session_id)
        Session ->> Session_DB: delete()
        Session -->> Auth: success
        Auth -->> GW: success
        GW -->> Client: 204 No Content
        Auth -->> Broker: UserLoggedOut
        Broker -->> Audit: UserLoggedOut
        Audit ->> Audit: save audit
    end
```

## Смена пароля

```mermaid
sequenceDiagram
    autonumber
    actor Client
    participant GW as Gateway
    participant Auth
    participant Auth_DB
    participant Session
    participant Session_DB
    participant Broker
    participant Audit

    Client ->> GW: POST /change-password
    GW ->> Session: CheckSession()
    Session -->> GW: SessionResponse()

    alt no_session
        GW -->> Client: 302 redirect
    else has
        GW ->> Auth: ChangePassword(old, new)
        Auth ->> Auth: verify old password
        Auth ->> Auth: save new hash
        Auth ->> Auth_DB: store()
        Auth ->> Session: RevokeAnotherSessions(user_id, session_id)
        Session -->> Auth: success
        Auth -->> GW: success
        GW -->> Client: 200 OK
        Session -->> Broker: SessionClosed
        Broker -->> Audit: SessionClosed
        Auth -->> Broker: PasswordChanged
        Broker -->> Audit: PasswordChanged
    end
```
