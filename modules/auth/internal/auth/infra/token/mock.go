package token

import (
	"strconv"
	"sync"
	"time"
	"uuid"

	"auth/internal/auth/app"
)

type MockJWT struct {
	mu      sync.RWMutex
	tokens  map[string]app.Claims
	counter int
	ttl     time.Duration
}

func NewMockJWT(ttl time.Duration) *MockJWT {
	return &MockJWT{
		tokens: make(map[string]app.Claims),
		ttl:    ttl,
	}
}

func (m *MockJWT) Issue(accountID, sessionID uuid.UUID) (app.AccessToken, error) {
	now := time.Now().UTC()
	m.mu.Lock()
	m.counter++
	n := m.counter
	token := "mock_token_" + strconv.Itoa(n)
	m.tokens[token] = app.Claims{
		AccountID: accountID,
		SessionID: sessionID,
		ExpiresAt: now.Add(m.ttl),
	}
	m.mu.Unlock()
	return app.AccessToken{
		Token:     token,
		ExpiresAt: now.Add(m.ttl),
	}, nil
}

func (m *MockJWT) Parse(raw string) (app.Claims, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	claims, ok := m.tokens[raw]
	if !ok {
		return app.Claims{}, app.ErrInvalidToken
	}
	if time.Now().UTC().After(claims.ExpiresAt) {
		return app.Claims{}, app.ErrInvalidToken
	}
	return claims, nil
}
