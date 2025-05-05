package data_test

import (
	"testing"
	"time"

	"github.com/shortykevich/greenlight/internal/data"
	"github.com/shortykevich/greenlight/internal/testutils/assertions"
)

func TestTokenInMem(t *testing.T) {
	tokens := data.NewTokenInMemRepo(nil)

	tActiv, err := tokens.New(1, 1*time.Minute, data.ScopeActivation)
	assertions.AssertNoError(t, err)
	tAuth, err := tokens.New(1, 1*time.Minute, data.ScopeAuthentication)
	assertions.AssertNoError(t, err)

	got, _ := tokens.GetToken(data.ScopeActivation, tActiv.Hash)
	assertions.AssertNoError(t, err)
	assertions.AssertTokens(t, got, tActiv)

	err = tokens.DeleteAllForUser(data.ScopeActivation, 1)
	assertions.AssertNoError(t, err)

	_, exist := tokens.GetToken(data.ScopeActivation, tActiv.Hash)
	if exist {
		t.Error("didn't expect to get existing value")
	}

	_, exist = tokens.GetToken(data.ScopeAuthentication, tAuth.Hash)
	if !exist {
		t.Error("expected to get value")
	}
}
