package data_test

import (
	"testing"

	"github.com/shortykevich/greenlight/internal/data"
	"github.com/shortykevich/greenlight/internal/testutils/assertions"
)

func TestUsers(t *testing.T) {
	users := data.NewUserInMemRepo()

	u1 := newUser("alice", "alice@example.com", "pa55word")
	err := users.Insert(u1)
	assertions.AssertNoError(t, err)

	if !users.UserExists("alice@example.com") {
		t.Errorf("expected to find user with email: '%s' but did not", u1.Email)
	}

	err = users.Insert(u1)
	assertions.AssertDuplicateError(t, err)

	_, err = users.GetByEmail("bob@example.com")
	assertions.AssertNotFoundError(t, err)

	got, err := users.GetByEmail("alice@example.com")
	assertions.AssertNoError(t, err)
	if got != u1 {
		t.Errorf("expected: %+v but got: %+v", u1, got)
	}

	got.Email = "alicenew@example.com"
	err = users.Update(got)
	assertions.AssertNoError(t, err)
	_, err = users.GetByEmail("alicenew@example.com")
	assertions.AssertNoError(t, err)
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
