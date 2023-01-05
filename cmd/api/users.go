package main

import (
	"easylist/internal/data"
	"easylist/internal/validator"
	"errors"
	"github.com/octoper/go-ray"
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

	if app.config.confirmation {
		token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		app.background(func() {
			var data = map[string]any{
				"activationToken": token.Plaintext,
				"domain":          app.config.domain,
			}
			err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
			if err != nil {
				app.logger.PrintError(err, nil)
			}
		})
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

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	type attributes struct {
		Token string `json:"token"`
	}
	type inputAttributes struct {
		Type       string     `json:"type"`
		Attributes attributes `json:"attributes"`
	}
	var input struct {
		Data inputAttributes `json:"data"`
	}
	var v = validator.New()

	var err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, "activateUserHandler", err)
		return
	}
	v.Check(input.Data.Type == "tokens", "data.type", "Wrong type provided, accepted type is tokens")

	if data.ValidateTokenPlaintext(v, input.Data.Attributes.Token); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.Data.Attributes.Token)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("data.attributes.token", "Invalid or expired activation token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	user.IsActive = true
	ray.Ray(user)
	err = app.models.Users.Update(user)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r, "activateUserHandler")
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{
		Id:         user.ID,
		TypeData:   USERS_TYPE_NAME,
		Attributes: user,
	}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
