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

func createTestList(app *application, list *data.List) error {
	if list.Name == "" {
		list.Name = "Test List"
	}
	if list.FolderId == 0 {
		list.FolderId = 1
	}
	if list.Order == 0 {
		list.Order = 1
	}
	if list.Version == 0 {
		list.Version = 1
	}
	if list.Icon == "" {
		list.Icon = "fa-list"
	}
	if list.UserId == 0 {
		list.UserId = 1
	}
	var err = app.models.Lists.Insert(list)
	if err != nil {
		return err
	}
	return nil
}

func createTestItem(app *application, item *data.Item) error {
	if item.Name == "" {
		item.Name = "Test Item"
	}
	if item.Description == "" {
		item.Description = "This is test Description"
	}
	if item.Quantity == 0 {
		item.Quantity = 1
	}
	if item.QuantityType == "" {
		item.QuantityType = "st"
	}
	if item.Price == 0 {
		item.Price = 78
	}
	if item.Order == 0 {
		item.Order = 1
	}
	var err = app.models.Items.Insert(item)
	if err != nil {
		return err
	}
	return nil
}
