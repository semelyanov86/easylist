package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"
)

const ScopeActivation = "activation"

type TokenModel struct {
	DB *sql.DB
}

type Token struct {
	Plaintext string
	Hash      []byte
	UserId    int64
	Expiry    time.Time
	Scope     string
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
	token.Hash = hash[:]
	return token, nil
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
