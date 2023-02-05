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
	if !app.config.Registration {
		app.methodNotAllowedResponse(w, r)
		return
	}

	var input = Input[UserAttributes]{Data: InputAttributes[UserAttributes]{
		Type:       "tokens",
		Attributes: UserAttributes{},
	}}

	var err = readJSON(w, r, &input)
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
	if app.config.Confirmation {
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
	err = app.models.Permissions.AddForUser(user.ID, "folders:read", "folders:write", "lists:write", "lists:read", "items:read", "items:write")
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if app.config.Confirmation {
		token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		app.background(func() {
			var data = map[string]any{
				"activationToken": token.Plaintext,
				"Domain":          app.config.Domain,
			}
			err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
			if err != nil {
				app.logger.PrintError(err, nil)
			}
		})
	}

	err = app.writeJSON(w, http.StatusCreated, user, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input = Input[ActivationAttributes]{Data: InputAttributes[ActivationAttributes]{
		Type:       "tokens",
		Attributes: ActivationAttributes{},
	}}

	var v = validator.New()

	var err = readJSON(w, r, &input)
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

	err = app.writeJSON(w, http.StatusOK, user, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
