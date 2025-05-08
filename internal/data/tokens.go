package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"slices"
	"sync"
	"time"

	"github.com/shortykevich/greenlight/internal/validator"
)

const (
	MockToken = "abcdefghijklmnopqrstuvwxyz"

	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)

type TokenRepository interface {
	TokenWriter
}

type TokenReader interface {
	GetToken(scope string, hash []byte) (*Token, bool)
}

type TokenWriter interface {
	New(userID int64, ttl time.Duration, scope string) (*Token, error)
	Insert(token *Token) error
	DeleteAllForUser(scope string, userID int64) error
}

type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserID    int64     `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, err
	}

	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}

func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(len(tokenPlaintext) > 0, "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

type TokenModel struct {
	DB *sql.DB
}

func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)
	return token, err
}

func (m TokenModel) Insert(token *Token) error {
	query := `
		INSERT INTO tokens (hash, user_id, expiry, scope)
		VALUES ($1, $2, $3, $4)`

	args := []any{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {
	query := `
		DELETE FROM tokens
		WHERE scope = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	return err
}

type TokenInMemRepo struct {
	mu     sync.RWMutex
	tokens map[string]*Token
	users  UserReader
}

func NewTokenInMemRepo(users UserReader) *TokenInMemRepo {
	return &TokenInMemRepo{
		tokens: make(map[string]*Token),
		users:  users,
	}
}

func generateMockToken(userID int64, _ time.Duration, scope string) *Token {
	hash := sha256.Sum256([]byte("abcdefghijklmnopqrstuvwxyz"))
	return &Token{
		UserID:    userID,
		Expiry:    MockTimeStamp,
		Scope:     scope,
		Plaintext: "abcdefghijklmnopqrstuvwxyz",
		Hash:      hash[:],
	}
}

func (m *TokenInMemRepo) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token := generateMockToken(userID, ttl, scope)

	err := m.Insert(token)
	return token, err
}

func (m *TokenInMemRepo) Insert(token *Token) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tokens[string(token.Hash)] = token
	return nil
}

func (m *TokenInMemRepo) DeleteAllForUser(scope string, userID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for k, v := range m.tokens {
		if v.Scope == scope && v.UserID == userID {
			delete(m.tokens, k)
		}
	}

	return nil
}

func (m *TokenInMemRepo) GetToken(scope string, hash []byte) (*Token, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, t := range m.tokens {
		hashMatches := slices.Equal(t.Hash, hash[:])
		notExpired := time.Now().Before(t.Expiry)
		if hashMatches && notExpired && t.Scope == scope {
			return t, true
		}
	}
	return nil, false
}
