package main

import (
	"easylist/internal/data"
	"easylist/internal/validator"
	"errors"
	"net/http"
	"time"
)

func (app *application) createActivationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, "createActivationTokenHandler", err)
		return
	}

	v := validator.New()

	if data.ValidateEmail(v, input.Email); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("email", "no matching email address found")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if user.IsActive {
		v.AddError("email", "user has already been activated")
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.background(func() {
		data := map[string]interface{}{
			"activationToken": token.Plaintext,
			"domain":          app.config.domain,
		}

		err = app.mailer.Send(user.Email, "token_activation.tmpl", data)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})
	var env = envelope{
		Id:       0,
		TypeData: "tokens",
		Attributes: map[string]any{
			"message": "an email will be sent to you containing activation instructions",
		},
	}

	err = app.writeJSON(w, http.StatusAccepted, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	type attributes struct {
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
		app.badRequestResponse(w, r, "createAuthenticationTokenHandler", err)
		return
	}
	var v = validator.New()
	v.Check(input.Data.Type == "tokens", "data.type", "Wrong type provided, accepted type is tokens")
	data.ValidateEmail(v, input.Data.Attributes.Email)
	data.ValidatePasswordPlaintext(v, input.Data.Attributes.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	user, err := app.models.Users.GetByEmail(input.Data.Attributes.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	match := user.Password.Matches(input.Data.Attributes.Password)
	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}
	token, err := app.models.Tokens.New(user.ID, 24*time.Hour*90, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusCreated, envelope{
		TypeData:   "tokens",
		Attributes: token,
	}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
