package data

import (
	"context"
	"database/sql"
	"easylist/internal/validator"
	"errors"
	"github.com/go-sql-driver/mysql"
	"github.com/jameskeane/bcrypt"
	_ "github.com/jameskeane/bcrypt"
	"strings"
	"time"
)

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	IsActive  bool      `json:"is_active"`
	Version   int       `json:"-"`
}

type password struct {
	plaintext *string
	hash      []byte
}

var ErrDuplicateEmail = errors.New("duplicate email")

type UserModel struct {
	DB *sql.DB
}

func (p *password) Set(plaintext string) error {
	hash, err := bcrypt.HashBytes([]byte(plaintext))
	if err != nil {
		return err
	}

	p.plaintext = &plaintext
	p.hash = hash

	return nil
}

func (p *password) Matches(plaintext string) bool {
	return bcrypt.MatchBytes([]byte(plaintext), p.hash)
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "data.attributes.email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) > 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) < 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) < 190, "name", "must not be more then 190 bytes")
	ValidateEmail(v, user.Email)
	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

func (u UserModel) Insert(user *User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Version = 1

	var query = `
				INSERT INTO users (name, email, password, is_active, created_at, updated_at, version) 
				VALUES (?, ?, ?, ?, ?, ?, ?)`
	var args = []any{user.Name, user.Email, user.Password.hash, user.IsActive, user.CreatedAt, user.UpdatedAt, user.Version}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := u.DB.ExecContext(ctx, query, args...)
	if err != nil {
		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError) {
			if mySQLError.Number == 1062 && strings.Contains(mySQLError.Message, "users.email") {
				return ErrDuplicateEmail
			}
		}
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = id

	return nil
}

func (u UserModel) GetByEmail(email string) (*User, error) {
	var query = `SELECT id, name, email, password, created_at, updated_at, is_active, version FROM users WHERE email = ?`
	var user User
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := u.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
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
	var query = `UPDATE users SET name = ?, email = ?, password = ?, is_active = ?, version = version + 1, updated_at = NOW() WHERE id = ? AND version = ?`
	var args = []any{user.Name, user.Email, user.Password.hash, user.IsActive, user.ID, user.Version}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := u.DB.QueryRowContext(ctx, query, args...).Scan()
	if err != nil {
		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError) {
			if mySQLError.Number == 1062 && strings.Contains(mySQLError.Message, "users.email") {
				return ErrDuplicateEmail
			}
		}
		if errors.Is(err, sql.ErrNoRows) {
			return ErrEditConflict
		}
		return err
	}
	user.Version++
	return nil
}
