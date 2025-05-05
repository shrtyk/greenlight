package assertions

import (
	"reflect"
	"slices"
	"testing"

	"github.com/shortykevich/greenlight/internal/data"
)

func AssertNoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("didn't expect error but got one: %v", err)
	}
}

func AssertDuplicateError(t testing.TB, err error) {
	t.Helper()
	if err != data.ErrDuplicateEmail {
		t.Error("expected duplicate error but didn't get one")
	}
}

func AssertNotFoundError(t testing.TB, err error) {
	t.Helper()
	if err != data.ErrRecordNotFound {
		t.Error("expected not found error but didn't get one")
	}
}

func AssertStatusCode(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("got status %v, want status %v", got, want)
	}
}

func AssertStrings(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("\ngot:\n%v\nwant:\n%v", got, want)
	}
}

func AssertTokens(t testing.TB, got, want *data.Token) {
	t.Helper()
	if !slices.Equal(got.Hash, want.Hash) {
		t.Errorf("got: %s, want: %s", got.Hash, want.Hash)
	}
}

func AssertUsers(t testing.TB, got, want *data.User) {
	t.Helper()
	if !reflect.DeepEqual(*got, *want) {
		t.Errorf("got: %v, want: %v", *got, *want)
	}
}

func AsserPermissions(t testing.TB, got, want data.Permissions) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("got: %v, wnat: %v", got, want)
	}
}
