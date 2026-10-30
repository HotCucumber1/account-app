package service

import (
	"context"
	"testing"
	"time"
	"uuid"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"session/internal/session/domain"
	"session/internal/session/domain/model"
)

type memorySessionRepository struct {
	sessions map[uuid.UUID]model.Session
}

func (r *memorySessionRepository) NextID() uuid.UUID {
	return uuid.NewV7()
}

func (r *memorySessionRepository) Store(_ context.Context, session model.Session) error {
	r.sessions[session.ID()] = session
	return nil
}

func (r *memorySessionRepository) Update(ctx context.Context, session model.Session) error {
	return r.Store(ctx, session)
}

func (r *memorySessionRepository) FindByID(_ context.Context, id uuid.UUID) (model.Session, error) {
	session, ok := r.sessions[id]
	if !ok {
		return model.Session{}, errors.WithStack(domain.ErrSessionNotFound)
	}
	return session, nil
}

func (r *memorySessionRepository) FindActiveByAccount(_ context.Context, accountID uuid.UUID, now time.Time) ([]model.Session, error) {
	var sessions []model.Session
	for _, session := range r.sessions {
		if session.Device().AccountID() == accountID && session.RevokedAt().IsZero() && now.Before(session.ExpiresAt()) {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

type memoryDeviceRepository struct {
	devices []model.Device
}

func (r *memoryDeviceRepository) NextID() uuid.UUID {
	return uuid.NewV7()
}

func (r *memoryDeviceRepository) Store(_ context.Context, device model.Device) error {
	r.devices = append(r.devices, device)
	return nil
}

func (r *memoryDeviceRepository) Find(_ context.Context, accountID uuid.UUID, userAgent string) (model.Device, error) {
	for _, device := range r.devices {
		if device.AccountID() == accountID && device.UserAgent() == userAgent {
			return device, nil
		}
	}
	return model.Device{}, errors.WithStack(domain.ErrDeviceNotFound)
}

type memoryDispatcher struct {
	events []model.Event
}

func (d *memoryDispatcher) Dispatch(_ context.Context, event model.Event) error {
	d.events = append(d.events, event)
	return nil
}

func TestSessionLifecycle(t *testing.T) {
	// Arrange
	sessions := &memorySessionRepository{sessions: map[uuid.UUID]model.Session{}}
	service := NewSessionService(sessions, &memoryDeviceRepository{}, &memoryDispatcher{}, time.Hour)
	accountID := uuid.New()

	// Act
	first, err := service.Create(t.Context(), accountID, "Firefox", "::ffff:10.0.0.1")
	// Assert
	require.NoError(t, err)
	require.True(t, first.NewDevice)
	require.Equal(t, "10.0.0.1", first.Session.IPAddress())

	// Act
	second, err := service.Create(t.Context(), accountID, "Firefox", "10.0.0.2")
	// Assert
	require.NoError(t, err)
	require.False(t, second.NewDevice)

	// Act
	killed, err := service.KillAccountSessions(t.Context(), accountID, first.Session.ID(), string(model.ReasonPasswordChanged))
	// Assert
	require.NoError(t, err)
	require.Equal(t, 1, killed)

	// Act
	require.NoError(t, service.Revoke(t.Context(), first.Session.ID(), string(model.ReasonLogout)))
	// Assert
	s := sessions.sessions[first.Session.ID()]
	require.Equal(t, model.ReasonLogout, s.RevokeReason())

	// Act
	_, err = service.Create(t.Context(), accountID, "Firefox", "not-ip")
	// Assert
	require.ErrorIs(t, err, domain.ErrInvalidSessionParams)

	// Act
	require.ErrorIs(t, service.Revoke(t.Context(), first.Session.ID(), "unknown"), domain.ErrInvalidRevokeReason)
}
