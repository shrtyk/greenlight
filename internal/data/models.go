package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

// Wrapper for all API models
type Models struct {
	Movies MovieRepository
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
	}
}

func NewMockModels() Models {
	return Models{
		Movies: &MovieInMemRepo{
			idCounter: 1,
			movies:    make(map[int64]*Movie),
		}}
}
