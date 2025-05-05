package data_test

import (
	"testing"

	"github.com/shortykevich/greenlight/internal/data"
	"github.com/shortykevich/greenlight/internal/testutils/assertions"
)

func TestPerms(t *testing.T) {
	perms := data.NewPermissionInMemRepo(nil)

	full := data.Permissions{"movies:read", "movies:write"}
	read := data.Permissions{"movies:read"}
	write := data.Permissions{"movies:write"}

	perms.AddForUser(1, full...)
	perms.AddForUser(2, read...)
	perms.AddForUser(3, write...)

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
