package service

import (
	"context"
	"net/netip"
	"time"
	"unicode/utf8"
	"uuid"

	"github.com/pkg/errors"

	"session/internal/session/domain"
	"session/internal/session/domain/model"
	"session/internal/session/domain/repo"
)

const maxUserAgentLength = 1024

type CreatedSession struct {
	Session   model.Session
	NewDevice bool
}

type SessionService struct {
	sessionRepository repo.SessionRepository
	deviceRepository  repo.DeviceRepository
	dispatcher        EventDispatcher
	ttl               time.Duration
}

func NewSessionService(
	sessionRepository repo.SessionRepository,
	deviceRepository repo.DeviceRepository,
	dispatcher EventDispatcher,
	ttl time.Duration,
) *SessionService {
	return &SessionService{
		sessionRepository: sessionRepository,
		deviceRepository:  deviceRepository,
		dispatcher:        dispatcher,
		ttl:               ttl,
	}
}

func (s *SessionService) Create(ctx context.Context, accountID uuid.UUID, userAgent, ipAddress string) (CreatedSession, error) {
	address, err := netip.ParseAddr(ipAddress)
	if err != nil || utf8.RuneCountInString(userAgent) > maxUserAgentLength {
		return CreatedSession{}, errors.WithStack(domain.ErrInvalidSessionParams)
	}
	now := time.Now().UTC()
	device, newDevice, err := s.resolveDevice(ctx, accountID, userAgent, now)
	if err != nil {
		return CreatedSession{}, err
	}
	session := model.NewSession(
		s.sessionRepository.NextID(),
		device,
		address.Unmap().String(),
		now,
		now.Add(s.ttl),
	)
	err = s.sessionRepository.Store(ctx, session)
	if err != nil {
		return CreatedSession{}, err
	}
	err = s.dispatcher.Dispatch(ctx, model.SessionCreated{
		SessionID: session.ID(),
		AccountID: accountID,
		DeviceID:  device.ID(),
		IPAddress: session.IPAddress(),
		NewDevice: newDevice,
	})
	if err != nil {
		return CreatedSession{}, err
	}
	return CreatedSession{
		Session:   session,
		NewDevice: newDevice,
	}, nil
}

func (s *SessionService) Revoke(ctx context.Context, id uuid.UUID, reason string) error {
	revokeReason, err := s.parseReason(reason)
	if err != nil {
		return err
	}
	session, err := s.sessionRepository.FindByID(ctx, id)
	if err != nil {
		return err
	}
	return s.revoke(ctx, session, revokeReason)
}

func (s *SessionService) KillAccountSessions(ctx context.Context, accountID, exceptSessionID uuid.UUID, reason string) (int, error) {
	revokeReason, err := s.parseReason(reason)
	if err != nil {
		return 0, err
	}
	sessions, err := s.sessionRepository.FindActiveByAccount(ctx, accountID, time.Now())
	if err != nil {
		return 0, err
	}
	killed := 0
	for _, session := range sessions {
		if session.ID() == exceptSessionID {
			continue
		}
		err = s.revoke(ctx, session, revokeReason)
		if err != nil {
			return killed, err
		}
		killed++
	}
	return killed, nil
}

func (s *SessionService) revoke(ctx context.Context, session model.Session, reason model.RevokeReason) error {
	now := time.Now().UTC()
	if !s.isActive(session, now) {
		return nil
	}

	session.SetRevokedAt(now)
	session.SetRevokeReason(reason)

	err := s.sessionRepository.Update(ctx, session)
	if err != nil {
		return err
	}

	return s.dispatcher.Dispatch(ctx, model.SessionClosed{
		SessionID: session.ID(),
		AccountID: session.Device().AccountID(),
		Reason:    reason,
	})
}

func (s *SessionService) isActive(session model.Session, now time.Time) bool {
	return session.RevokedAt().IsZero() && now.Before(session.ExpiresAt())
}

func (s *SessionService) parseReason(value string) (model.RevokeReason, error) {
	reason := model.RevokeReason(value)
	switch reason {
	case model.ReasonLogout,
		model.ReasonPasswordChanged,
		model.ReasonRevoked:
		return reason, nil
	default:
		return "", errors.WithStack(domain.ErrInvalidRevokeReason)
	}
}

func (s *SessionService) resolveDevice(ctx context.Context, accountID uuid.UUID, userAgent string, now time.Time) (model.Device, bool, error) {
	device, err := s.deviceRepository.Find(ctx, accountID, userAgent)
	if err == nil {
		return device, false, nil
	}
	if !errors.Is(err, domain.ErrDeviceNotFound) {
		return model.Device{}, false, err
	}

	device = model.NewDevice(s.deviceRepository.NextID(), accountID, userAgent, now)
	err = s.deviceRepository.Store(ctx, device)

	if errors.Is(err, domain.ErrDeviceAlreadyExists) {
		device, err = s.deviceRepository.Find(ctx, accountID, userAgent)
		return device, false, err
	}
	return device, err == nil, err
}
