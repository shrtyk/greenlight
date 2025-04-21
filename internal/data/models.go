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

// Wrapper for all API models
type Models struct {
	Movies MovieRepository
	Users  UserRepository
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
		Users:  UserModel{DB: db},
	}
}

func NewMockModels() Models {
	return Models{
		Movies: &MovieInMemRepo{
			idCounter: 1,
			movies:    make(map[int64]*Movie),
		},
		Users: &UserInMemRepo{
			idCounter: 1,
			users:     make(map[int64]*User),
		},
	}
}
