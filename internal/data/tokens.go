package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"sync"
	"time"

	"github.com/shrtyk/greenlight/internal/validator"
)

const (
	MockToken = "abcdefghijklmnopqrstuvwxyz"

	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)

type TokenRepository interface {
	TokenWriter
	TokenReader
}

type TokenReader interface {
	GetToken(scope string, hash []byte) (*Token, bool)
	GetUserTokens(userID int64) *Tokens
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

func (m TokenModel) GetToken(scope string, hash []byte) (*Token, bool) {
	return nil, false
}

func (m TokenModel) GetUserTokens(userID int64) *Tokens {
	return nil
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

func (m *TokenInMemRepo) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	token.Expiry = MockTimeStamp
	err = m.Insert(token)
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
	t, ok := m.tokens[string(hash)]
	m.mu.RUnlock()

	if !ok || t.Scope != scope || time.Now().After(t.Expiry) {
		return nil, false
	}
	return t, true
}

type Tokens struct {
	Activation     *string
	Authentication *string
}

func (m *TokenInMemRepo) GetUserTokens(userID int64) *Tokens {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tokens := new(Tokens)
	for _, v := range m.tokens {
		if v.UserID != userID {
			continue
		}
		t := v.Plaintext
		switch v.Scope {
		case ScopeActivation:
			tokens.Activation = &t
		case ScopeAuthentication:
			tokens.Authentication = &t
		}
	}
	return tokens
}
