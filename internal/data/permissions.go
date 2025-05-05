package data

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"sync"
	"time"
)

const (
	MoviesWrite = "movies:write"
	MoviesRead  = "movies:read"
)

type PermissionRepository interface {
	GetAllForUser(userID int64) (permissions Permissions, err error)
	AddForUser(userID int64, codes ...string) error
}

type Permissions []string

func (p Permissions) Include(code string) bool {
	return slices.Contains(p, code)
}

type PermissionModel struct {
	DB *sql.DB
}

func (m PermissionModel) GetAllForUser(userID int64) (permissions Permissions, err error) {
	query := `
		SELECT permissions.code
		FROM permissions
		INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id
		INNER JOIN users ON users_permissions.user_id = users.id
		WHERE users.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errors.Join(err, rows.Close())
	}()

	for rows.Next() {
		var permission string
		if err = rows.Scan(&permission); err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return
}

func (m PermissionModel) AddForUser(userID int64, codes ...string) error {
	query := `
		INSERT INTO users_permissions
		SELECT $1, permissions.id FROM permissions WHERE permissions.code = ANY($2)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, userID, codes)
	return err
}

type PermissionInMemRepo struct {
	mu          sync.RWMutex
	permissions map[int]string
	users       UserReader
}

func NewPermissionInMemRepo(users UserReader) *PermissionInMemRepo {
	return &PermissionInMemRepo{
		permissions: make(map[int]string),
		users:       users,
	}
}

func (m *PermissionInMemRepo) GetAllForUser(userID int64) (permissions Permissions, err error) {
	// TODO
	return nil, nil
}

func (m *PermissionInMemRepo) AddForUser(userID int64, codes ...string) error {
	// TODO
	return nil
}
