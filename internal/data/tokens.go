package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"easylist/internal/validator"
	"encoding/base32"
	"encoding/hex"
	_ "github.com/octoper/go-ray"
	"time"
)

const ScopeActivation = "activation"
const ScopeAuthentication = "authentication"

type TokenModel struct {
	DB *sql.DB
}

type Token struct {
	Plaintext string    `json:"token"`
	Hash      string    `json:"-"`
	UserId    int64     `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

func generateToken(userId int64, ttl time.Duration, scope string) (*Token, error) {
	var token = &Token{
		UserId: userId,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}
	var randomBytes = make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	var hash = sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hex.EncodeToString(hash[:])
	return token, nil
}

func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

func (t TokenModel) New(userId int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userId, ttl, scope)
	if err != nil {
		return nil, err
	}
	err = t.Insert(token)
	return token, err
}

func (t TokenModel) Insert(token *Token) error {
	var query = `INSERT INTO tokens (hash, user_id, expired_at, scope) VALUES (?,?,?,?)`
	var args = []any{token.Hash, token.UserId, token.Expiry, token.Scope}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := t.DB.ExecContext(ctx, query, args...)
	return err
}

func (t TokenModel) DeleteAllForUser(scope string, userId int64) error {
	var query = `DELETE FROM tokens WHERE scope = ? AND user_id = ?`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := t.DB.ExecContext(ctx, query, scope, userId)
	return err
}

type MockTokenModel struct {
}

func (t MockTokenModel) New(userId int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userId, ttl, scope)
	return token, err
}

func (t MockTokenModel) Insert(token *Token) error {
	return nil
}

func (t MockTokenModel) DeleteAllForUser(scope string, userId int64) error {
	return nil
}
