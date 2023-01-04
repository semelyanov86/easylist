package main

import (
	"easylist/internal/data"
	"easylist/internal/validator"
	"errors"
	"net/http"
	"time"
)

const USERS_TYPE_NAME = "users"

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	if !app.config.registration {
		app.methodNotAllowedResponse(w, r)
		return
	}

	type attributes struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type inputAttributes struct {
		Type       string     `json:"type"`
		Attributes attributes `json:"attributes"`
	}
	var input struct {
		Data inputAttributes `json:"data"`
	}

	var err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, "registerUserHandler", err)
		return
	}

	var user = &data.User{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      input.Data.Attributes.Name,
		Email:     input.Data.Attributes.Email,
		IsActive:  true,
		Version:   1,
	}
	if app.config.confirmation {
		user.IsActive = false
	}

	err = user.Password.Set(input.Data.Attributes.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	var v = validator.New()
	v.Check(input.Data.Type == USERS_TYPE_NAME, "data.type", "Wrong type provided, accepted type is users")

	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("data.attributes.email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{
		Id:         user.ID,
		TypeData:   USERS_TYPE_NAME,
		Attributes: user,
	}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
