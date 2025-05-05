package data_test

import (
	"testing"
	"time"

	"github.com/shortykevich/greenlight/internal/data"
	"github.com/shortykevich/greenlight/internal/testutils/assertions"
)

func TestUserInMem(t *testing.T) {
	models := data.NewMockModels()
	users, tokens := models.Users, models.Tokens

	u1 := newUser("alice", "alice@example.com", "pa55word")
	err := users.Insert(u1)
	assertions.AssertNoError(t, err)

	err = users.Insert(u1)
	assertions.AssertDuplicateError(t, err)

	_, err = users.GetByEmail("bob@example.com")
	assertions.AssertNotFoundError(t, err)

	alice, err := users.GetByEmail("alice@example.com")
	assertions.AssertNoError(t, err)
	assertions.AssertUsers(t, alice, u1)

	tkn, err := tokens.New(alice.ID, 1*time.Minute, data.ScopeActivation)
	assertions.AssertNoError(t, err)

	usr, err := users.GetForToken(data.ScopeActivation, tkn.Plaintext)
	assertions.AssertNoError(t, err)
	assertions.AssertUsers(t, usr, alice)

	alice.Email = "alicenew@example.com"
	err = users.Update(alice)
	assertions.AssertNoError(t, err)
	_, err = users.GetByEmail("alicenew@example.com")
	assertions.AssertNoError(t, err)

	_, err = users.GetForToken(data.ScopeActivation, "abcde")
	assertions.AssertNotFoundError(t, err)
}

func newUser(name, email, plainPassword string) *data.User {
	password := data.Password{}
	password.Set(plainPassword)
	return &data.User{
		Name:     name,
		Email:    email,
		Password: password,
	}
}
