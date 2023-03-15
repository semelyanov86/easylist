package main

import (
	"easylist/internal/data"
	"easylist/internal/validator"
	"errors"
	"net/http"
	"strconv"
	"time"
)

const USERS_TYPE_NAME = "users"

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	if !app.config.Registration {
		app.methodNotAllowedResponse(w, r)
		return
	}

	var input = Input[UserAttributes]{Data: InputAttributes[UserAttributes]{
		Type:       "users",
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

func (app *application) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	var userModel = app.contextGetUser(r)

	if id != userModel.ID {
		app.notPermittedResponse(w, r)
	}

	var input = Input[UserAttributes]{Data: InputAttributes[UserAttributes]{
		Type:       "users",
		Attributes: UserAttributes{},
	}}

	err = readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, "updateUserHandler", err)
		return
	}

	if input.Data.Attributes.Name != "" {
		userModel.Name = input.Data.Attributes.Name
	}
	if input.Data.Attributes.Email != "" {
		userModel.Email = input.Data.Attributes.Email
	}
	if input.Data.Attributes.Password != "" {
		err = userModel.Password.Set(input.Data.Attributes.Password)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	var v = validator.New()

	v.Check(input.Data.Type == USERS_TYPE_NAME, "data.type", "Wrong type provided, accepted type is users")
	if data.ValidateUser(v, userModel); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Users.Update(userModel)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r, "updateUserHandler")
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("data.attributes.email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if r.Header.Get("X-Expected-Version") != "" {
		if strconv.FormatInt(int64(userModel.Version), 32) != r.Header.Get("X-Expected-Version") {
			app.editConflictResponse(w, r, "updateUserHandler")
			return
		}
	}

	err = app.writeJSON(w, http.StatusOK, userModel, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	var userModel = app.contextGetUser(r)

	err := app.writeJSON(w, http.StatusOK, userModel, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	var userModel = app.contextGetUser(r)

	if userModel.ID != id {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Items.DeleteByUser(id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	err = app.models.Lists.DeleteByUser(id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	err = app.models.Folders.DeleteByUser(id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	err = app.models.Users.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) resetUserPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var input = Input[ResetPasswordAttributes]{Data: InputAttributes[ResetPasswordAttributes]{
		Type:       "tokens",
		Attributes: ResetPasswordAttributes{},
	}}

	err := readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, "resetUserPasswordHandler", err)
		return
	}

	v := validator.New()

	data.ValidatePasswordPlaintext(v, input.Data.Attributes.Password)
	data.ValidateTokenPlaintext(v, input.Data.Attributes.Token)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetForToken(data.ScopePasswordReset, input.Data.Attributes.Token)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired password reset token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = user.Password.Set(input.Data.Attributes.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r, "resetUserPasswordHandler")
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.models.Tokens.DeleteAllForUser(data.ScopePasswordReset, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, user, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
