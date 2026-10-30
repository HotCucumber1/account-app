package password

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type Bcrypt struct {
	cost int
}

func NewBcrypt(cost int) *Bcrypt {
	return &Bcrypt{cost: cost}
}

func (b *Bcrypt) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	return string(hash), errors.WithStack(err)
}

func (b *Bcrypt) Verify(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
