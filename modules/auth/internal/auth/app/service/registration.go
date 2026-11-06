package service

import (
	"context"
	"errors"
	"log/slog"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/logging"
	pkgerrors "github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"auth/internal/auth/app"
	"auth/internal/auth/domain/model"
	"auth/internal/auth/domain/repo"
)

type RegistrationParticipant interface {
	Name() string
	Prepare(ctx context.Context, request app.RegistrationRequest) error
	Commit(ctx context.Context, transactionID string) error
	Rollback(ctx context.Context, transactionID string) error
}

type RegistrationConfig struct {
	Deadline      time.Duration
	RecoveryGrace time.Duration
	RecoveryBatch int
}

type RegistrationCoordinator struct {
	unitOfWork   UnitOfWork
	journal      repo.RegistrationRepository
	participants []RegistrationParticipant
	config       RegistrationConfig
	logger       *slog.Logger
}

func NewRegistrationCoordinator(
	unitOfWork UnitOfWork,
	journal repo.RegistrationRepository,
	participants []RegistrationParticipant,
	config RegistrationConfig,
	logger *slog.Logger,
) *RegistrationCoordinator {
	return &RegistrationCoordinator{
		unitOfWork:   unitOfWork,
		journal:      journal,
		participants: participants,
		config:       config,
		logger:       logger,
	}
}

func (c *RegistrationCoordinator) Register(ctx context.Context, input app.RegisterInput) (uuid.UUID, error) {
	request := app.RegistrationRequest{
		TransactionID: "reg-" + uuid.NewV7().String(),
		AccountID:     uuid.NewV7(),
		Email:         input.Email,
		Name:          input.Name,
		Password:      input.Password,
	}
	registration := model.NewRegistration(
		request.TransactionID,
		model.RegistrationPreparing,
		time.Now().UTC().Add(c.config.Deadline),
	)
	err := c.unitOfWork.Execute(ctx, func(ctx context.Context) error {
		return c.journal.Add(ctx, registration, request.AccountID)
	})
	if err != nil {
		return uuid.Nil(), err
	}

	prepareErr := c.prepare(ctx, registration.Deadline(), request)

	ctx = context.WithoutCancel(ctx)
	if prepareErr != nil {
		_ = c.finish(ctx, request.TransactionID, model.RegistrationPreparing, model.RegistrationAborting)
		return uuid.Nil(), prepareErr
	}
	decided, err := c.transition(ctx, request.TransactionID, model.RegistrationPreparing, model.RegistrationCommitting)
	if err != nil {
		return uuid.Nil(), err
	}
	if !decided {
		return uuid.Nil(), pkgerrors.WithStack(app.ErrRegistrationAborted)
	}
	_ = c.finish(ctx, request.TransactionID, model.RegistrationCommitting, model.RegistrationCommitting)
	return request.AccountID, nil
}

func (c *RegistrationCoordinator) Recover(ctx context.Context) error {
	var registrations []model.Registration
	err := c.unitOfWork.Execute(ctx, func(ctx context.Context) error {
		var err error
		registrations, err = c.journal.FindUnfinished(ctx, c.config.RecoveryBatch)
		return err
	})
	if err != nil {
		return err
	}
	var errs []error
	for _, registration := range registrations {
		switch registration.Status() {
		case model.RegistrationPreparing:
			if time.Now().UTC().After(registration.Deadline().Add(c.config.RecoveryGrace)) {
				errs = append(errs, c.finish(ctx, registration.TransactionID(), model.RegistrationPreparing, model.RegistrationAborting))
			}
		case model.RegistrationCommitting,
			model.RegistrationAborting:
			errs = append(errs, c.finish(ctx, registration.TransactionID(), registration.Status(), registration.Status()))
		}
	}
	return errors.Join(errs...)
}

func (c *RegistrationCoordinator) prepare(ctx context.Context, deadline time.Time, request app.RegistrationRequest) error {
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()
	group, ctx := errgroup.WithContext(ctx)
	for _, participant := range c.participants {
		group.Go(func() error {
			return pkgerrors.Wrapf(participant.Prepare(ctx, request), "prepare %s", participant.Name())
		})
	}
	return group.Wait()
}

func (c *RegistrationCoordinator) finish(ctx context.Context, transactionID string, from, decision model.RegistrationStatus) error {
	if from != decision {
		decided, err := c.transition(ctx, transactionID, from, decision)
		if err != nil || !decided {
			return err
		}
	}
	final := model.RegistrationAborted
	deliver := RegistrationParticipant.Rollback
	if decision == model.RegistrationCommitting {
		final = model.RegistrationCommitted
		deliver = RegistrationParticipant.Commit
	}
	var errs []error
	for _, participant := range c.participants {
		err := deliver(participant, ctx, transactionID)
		if err != nil {
			errs = append(errs, pkgerrors.Wrapf(err, "%s %s", decision, participant.Name()))
		}
	}
	err := errors.Join(errs...)
	if err == nil {
		_, err = c.transition(ctx, transactionID, decision, final)
	}
	if err != nil {
		c.logger.WarnContext(ctx, "registration decision is not delivered yet; recovery will retry",
			slog.String("transaction_id", transactionID), slog.String("decision", string(decision)), logging.Error(err),
		)
	}
	return err
}

func (c *RegistrationCoordinator) transition(ctx context.Context, transactionID string, from, to model.RegistrationStatus) (bool, error) {
	var updated bool
	err := c.unitOfWork.Execute(ctx, func(ctx context.Context) error {
		var err error
		updated, err = c.journal.UpdateStatus(ctx, transactionID, from, to)
		return err
	})
	return updated, err
}
