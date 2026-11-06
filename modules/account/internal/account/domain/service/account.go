package service

import (
	"context"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"
	"uuid"

	"github.com/pkg/errors"

	"account/internal/account/domain"
	"account/internal/account/domain/model"
	"account/internal/account/domain/repo"
)

const maxNameLength = 255

type AccountService struct {
	repository repo.AccountRepository
	dispatcher EventDispatcher
}

func NewAccountService(repository repo.AccountRepository, dispatcher EventDispatcher) *AccountService {
	return &AccountService{
		repository: repository,
		dispatcher: dispatcher,
	}
}

func (s *AccountService) Create(ctx context.Context, email, name string) (model.Account, error) {
	return s.CreateWithID(ctx, s.repository.NextID(), email, name)
}

func (s *AccountService) CreateWithID(ctx context.Context, id uuid.UUID, email, name string) (model.Account, error) {
	email = normalizeEmail(email)
	name = strings.TrimSpace(name)
	err := s.validate(email, name)
	if err != nil {
		return model.Account{}, err
	}
	exists, err := s.repository.ExistsByEmail(ctx, email) // TODO email лучше в auth
	if err != nil {
		return model.Account{}, err
	}
	if exists {
		return model.Account{}, errors.WithStack(domain.ErrAccountAlreadyExists)
	}
	account := model.NewAccount(id, email, name, time.Now().UTC())
	err = s.repository.Store(ctx, account)
	if err != nil {
		return model.Account{}, err
	}
	err = s.dispatcher.Dispatch(ctx, model.AccountCreated{
		AccountID: account.ID(),
		Email:     account.Email(),
		Name:      account.Name(),
	})
	if err != nil {
		return model.Account{}, err
	}
	return account, nil
}

func (s *AccountService) Update(ctx context.Context, id uuid.UUID, name string) (model.Account, error) {
	name = strings.TrimSpace(name)
	err := validateName(name)
	if err != nil {
		return model.Account{}, err
	}
	account, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return model.Account{}, err
	}
	account.SetName(name)
	err = s.repository.Store(ctx, account)
	if err != nil {
		return model.Account{}, err
	}
	return account, nil
}

// TODO withStack на wrap (текст крч прокидывать)
func (s *AccountService) validate(email, name string) error {
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return errors.WithStack(domain.ErrInvalidAccount)
	}
	return validateName(name)
}

func validateName(name string) error {
	if name == "" || utf8.RuneCountInString(name) > maxNameLength {
		return errors.WithStack(domain.ErrInvalidAccount)
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
