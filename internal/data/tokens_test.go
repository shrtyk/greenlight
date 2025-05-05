package data_test

import (
	"testing"
	"time"

	"github.com/shortykevich/greenlight/internal/data"
	"github.com/shortykevich/greenlight/internal/testutils/assertions"
)

func TestTokenInMem(t *testing.T) {
	tokens := data.NewTokensInMemRepo()

	tActiv, err := tokens.New(1, 1*time.Minute, data.ScopeActivation)
	assertions.AssertNoError(t, err)
	tAuth, err := tokens.New(1, 1*time.Minute, data.ScopeAuthentication)
	assertions.AssertNoError(t, err)

	got, err := tokens.GetToken(tActiv.Hash)
	assertions.AssertNoError(t, err)
	assertions.AssertTokens(t, got, tActiv)

	err = tokens.DeleteAllForUser(data.ScopeActivation, 1)
	assertions.AssertNoError(t, err)

	_, err = tokens.GetToken(tActiv.Hash)
	assertions.AssertNotFoundError(t, err)

	_, err = tokens.GetToken(tAuth.Hash)
	assertions.AssertNoError(t, err)
}
