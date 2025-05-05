package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
	ErrCloseRows      = errors.New("couldn't close rows")
)

// Models is a wrapper for all API models.
type Models struct {
	Movies      MovieRepository
	Users       UserRepository
	Tokens      TokenRepository
	Permissions PermissionRepository
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies:      MovieModel{DB: db},
		Users:       UserModel{DB: db},
		Tokens:      TokenModel{DB: db},
		Permissions: PermissionModel{DB: db},
	}
}

func NewMockModels() Models {
	userRepo := &UserInMemRepo{
		users: make(map[int64]*User),
	}

	tokenRepo := NewTokenInMemRepo(userRepo)
	permRepo := NewPermissionInMemRepo(userRepo)

	userRepo.tokens = tokenRepo
	userRepo.permissions = permRepo

	return Models{
		Movies:      NewMovieInMemRepo(),
		Users:       userRepo,
		Tokens:      tokenRepo,
		Permissions: permRepo,
	}
}

func relateUsersAndTokens(users *UserInMemRepo, tokens *TokenInMemRepo) {
	tokens.users = users
	users.tokens = tokens
}

func relateUsersAndPermissions(users *UserInMemRepo, permissions *PermissionInMemRepo) {
	users.permissions = permissions
	permissions.users = users
}
