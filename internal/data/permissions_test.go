package data_test

import (
	"testing"

	"github.com/shrtyk/greenlight/internal/data"
	"github.com/shrtyk/greenlight/internal/testutils/assertions"
)

func TestPerms(t *testing.T) {
	perms := data.NewPermissionInMemRepo(nil)

	full := data.Permissions{"movies:read", "movies:write"}
	read := data.Permissions{"movies:read"}
	write := data.Permissions{"movies:write"}

	err := perms.AddForUser(1, full...)
	assertions.AssertNoError(t, err)
	_ = perms.AddForUser(2, read...)
	_ = perms.AddForUser(3, write...)

	u1perms, err := perms.GetAllForUser(1)
	assertions.AssertNoError(t, err)
	assertions.AssertPermissions(t, u1perms, full)

	u2perms, err := perms.GetAllForUser(2)
	assertions.AssertNoError(t, err)
	assertions.AssertPermissions(t, u2perms, read)

	u3perms, err := perms.GetAllForUser(3)
	assertions.AssertNoError(t, err)
	assertions.AssertPermissions(t, u3perms, write)
}
