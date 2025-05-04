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
	Permissions PermissionModel
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
	users, tokens := CreateRelatedUsersAndTokens()
	return Models{
		Movies: NewMovieInMemRepo(),
		Users:  users,
		Tokens: tokens,
	}
}

func CreateRelatedUsersAndTokens() (users *UserInMemRepo, tokens *TokenInMemRepo) {
	tokens, users = NewTokensInMemRepo(), NewUserInMemRepo()
	tokens.users = users
	users.tokens = tokens
	return
}
