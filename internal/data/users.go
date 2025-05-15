package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shrtyk/greenlight/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

var (
	AnonymousUser = &User{}
)

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  Password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

type Password struct {
	plaintext *string
	hash      []byte
}

func (p *Password) Set(plaintTextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintTextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintTextPassword
	p.hash = hash

	return nil
}

func (p *Password) Matches(plainTextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plainTextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func (u *User) Validate(v *validator.Validator) {
	v.Check(len(u.Name) > 0, "name", "must be provided")
	v.Check(len(u.Name) <= 500, "name", "must not be more than 500 bytes long")

	ValidateEmail(v, u.Email)

	if u.Password.plaintext != nil {
		ValidatePlainTextPassword(v, *u.Password.plaintext)
	}

	if u.Password.hash == nil {
		panic("missing password hash for user")
	}
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePlainTextPassword(v *validator.Validator, password string) {
	v.Check(len(password) > 0, "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

type UserRepository interface {
	UserReader
	UserWriter
}

type UserReader interface {
	GetByEmail(email string) (*User, error)
	GetForToken(scope, tokenPlaintext string) (*User, error)
}

type UserWriter interface {
	Insert(user *User) error
	Update(user *User) error
}

type UserModel struct {
	DB *sql.DB
}

func (u UserModel) Insert(user *User) error {
	query := `
		INSERT INTO users (name, email, password_hash, activated)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version`

	args := []any{user.Name, user.Email, user.Password.hash, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.As(err, &pgErr) && pgErr.Code == "23505":
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

func (u UserModel) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, created_at, name, email, password_hash, activated, version
		FROM users
		WHERE email = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (u UserModel) Update(user *User) error {
	query := `
		UPDATE users
		SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
		WHERE id = $5 and version = $6
		RETURNING version`

	args := []any{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.As(err, &pgErr) && pgErr.Code == "23505":
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (u UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
	query := `
		SELECT users.id, users.created_at, users.name, users.email, users.password_hash, users.activated, users.version
		FROM users
		INNER JOIN tokens
		ON users.id = tokens.user_id
		WHERE tokens.hash = $1
		AND tokens.scope = $2
		AND tokens.expiry > $3`

	tokenHash := sha256.Sum256([]byte(tokenPlaintext))
	args := []any{tokenHash[:], tokenScope, time.Now()}

	var user User
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

type UserInMemRepo struct {
	mu          sync.RWMutex
	idCounter   int64
	users       map[int64]*User
	tokens      TokenReader
	permissions PermissionRepository
	clock       Clock
}

func NewUserInMemRepo() *UserInMemRepo {
	return &UserInMemRepo{
		idCounter: 1,
		users:     make(map[int64]*User),
		clock:     MockClock{},
	}
}

func (m *UserInMemRepo) UserExists(email string) bool {
	for _, user := range m.users {
		if user.Email == email {
			return true
		}
	}
	return false
}

func (m *UserInMemRepo) GetByEmail(email string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, ErrRecordNotFound
}

func (m *UserInMemRepo) Insert(user *User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.UserExists(user.Email) {
		return ErrDuplicateEmail
	}

	user.ID = m.idCounter
	user.Version = 1
	user.CreatedAt = m.clock.Now()

	m.users[m.idCounter] = user
	m.idCounter++

	return nil
}

func (m *UserInMemRepo) Update(user *User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var u *User
	for _, us := range m.users {
		if us.ID == user.ID {
			u = us
		}
	}

	u.Name = user.Name
	u.Email = user.Email
	u.Password.hash = user.Password.hash
	u.Activated = user.Activated
	u.Version++

	return nil
}

func (m *UserInMemRepo) GetForToken(scope, tokenPlaintext string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tokenHash := sha256.Sum256([]byte(tokenPlaintext))
	t, exist := m.tokens.GetToken(scope, tokenHash[:])
	if !exist {
		return nil, ErrRecordNotFound
	}
	user, exist := m.users[t.UserID]
	if !exist {
		return nil, ErrRecordNotFound
	}

	return user, nil
}
