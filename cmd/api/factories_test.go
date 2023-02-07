package main

import (
	"database/sql"
	"easylist/internal/data"
	"testing"
	"time"
)

func createTestUserWithToken(t *testing.T, app *application, email string) (*data.User, *data.Token, error) {
	if email == "" {
		email = "test@mail.ru"
	}
	var token *data.Token
	var user = &data.User{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      "Test User",
		Email:     email,
		IsActive:  true,
		Version:   1,
	}

	err := user.Password.Set("password123")
	if err != nil {
		return user, token, err
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		return user, token, err
	}

	err = app.models.Permissions.AddForUser(user.ID, "folders:read", "folders:write", "lists:write", "lists:read", "items:read", "items:write")
	if err != nil {
		return user, token, err
	}
	token, err = app.models.Tokens.New(user.ID, 24*time.Hour*90, data.ScopeAuthentication)
	if err != nil {
		return user, token, err
	}
	return user, token, nil
}

func createTestFolder(app *application, userId int64, name string, order int32) (*data.Folder, error) {
	if name == "" {
		name = "Test Folder"
	}
	if order == 0 {
		order = 1
	}
	var folder = data.Folder{
		Name:    name,
		Icon:    "fa-folder",
		Version: 1,
		Order:   order,
		UserId: sql.NullInt64{
			Int64: userId,
			Valid: true,
		},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}
	var err = app.models.Folders.Insert(&folder)
	if err != nil {
		return &folder, err
	}
	return &folder, nil
}
