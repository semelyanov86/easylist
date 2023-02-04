package main

import (
	"database/sql"
	"easylist/internal/data"
	"testing"
	"time"
)

func createTestUserWithToken(t *testing.T, app *application) (*data.User, *data.Token, error) {
	var token *data.Token
	var user = &data.User{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      "Test User",
		Email:     "test@mail.ru",
		IsActive:  true,
		Version:   1,
	}

	err := user.Password.Set("password")
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

func createTestFolder(app *application, userId int64) (*data.Folder, error) {
	var folder = data.Folder{
		Name:    "Test Folder",
		Icon:    "fa-folder",
		Version: 1,
		Order:   1,
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
