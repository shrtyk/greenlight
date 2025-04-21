package data

import (
	"errors"
	"time"

	"github.com/shortykevich/greenlight/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID         int64     `json:"id"`
	Created_ad time.Time `json:"created_at"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Password   password  `json:"-"`
	Activated  bool      `json:"activated"`
	Version    int       `json:"-"`
}

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintTextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintTextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintTextPassword
	p.hash = hash

	return nil
}

func (p *password) Matches(plainTextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plainTextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func (u *User) Validate(v *validator.Validator) {
	v.Check(len(u.Name) > 0, "name", "must be provided")
	v.Check(len(u.Name) <= 500, "name", "must not be more than 500 bytes long")

	ValidateEmail(v, u.Email)

	if u.Password.plaintext != nil {
		ValidatePlainTextPassword(v, *u.Password.plaintext)
	}

	if u.Password.hash == nil {
		panic("missing password hash for user")
	}
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePlainTextPassword(v *validator.Validator, password string) {
	v.Check(len(password) > 0, "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}
