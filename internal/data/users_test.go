package data_test

import (
	"testing"

	"github.com/shortykevich/greenlight/internal/data"
)

func TestUsers(t *testing.T) {
	users := data.NewUserInMemRepo()

	u1 := newUser("alice", "alice@example.com", "pa55word")
	err := users.Insert(u1)
	assertError(t, err)

	if !users.UserExists("alice@example.com") {
		t.Errorf("expected to find user with email: '%s' but did not", u1.Email)
	}

	err = users.Insert(u1)
	assertDuplicateError(t, err)

	_, err = users.GetByEmail("bob@example.com")
	assertNotFoundError(t, err)

	got, err := users.GetByEmail("alice@example.com")
	assertError(t, err)
	if got != u1 {
		t.Errorf("expected: %+v but got: %+v", u1, got)
	}

	got.Email = "alicenew@example.com"
	err = users.Update(got)
	assertError(t, err)
	_, err = users.GetByEmail("alicenew@example.com")
	assertError(t, err)
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

func assertError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("didn't expect got an error: %v", err)
	}
}

func assertDuplicateError(t testing.TB, err error) {
	t.Helper()
	if err != data.ErrDuplicateEmail {
		t.Error("expected duplicate error but didn't get one")
	}
}

func assertNotFoundError(t testing.TB, err error) {
	t.Helper()
	if err != data.ErrRecordNotFound {
		t.Error("expected not found error but didn't get one")
	}
}
