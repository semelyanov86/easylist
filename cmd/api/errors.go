package main

import (
	"fmt"
	"github.com/google/jsonapi"
	"log"
	"net/http"
)

func (app *application) logError(r *http.Request, err error) {
	app.logger.PrintError(err, nil)
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message jsonapi.ErrorsPayload) {
	w.Header().Set("Content-Type", jsonapi.MediaType)
	w.WriteHeader(status)
	err := jsonapi.MarshalErrors(w, message.Errors)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)
	m := make(map[string]interface{})
	m["pointer"] = string(rune(log.Lshortfile))

	var errorObject = jsonapi.ErrorObject{
		Status: "500",
		Code:   "500",
		Meta:   &m,
		Title:  "Internal server error",
		Detail: "The server encountered a problem and could not process your request",
	}
	app.errorResponse(w, r, http.StatusInternalServerError, jsonapi.ErrorsPayload{Errors: []*jsonapi.ErrorObject{&errorObject}})
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	m := make(map[string]interface{})
	m["pointer"] = r.RequestURI

	var errorObject = jsonapi.ErrorObject{
		Status: "404",
		Code:   "404",
		Meta:   &m,
		Title:  "Not Found",
		Detail: "The requested resource could not be found",
	}
	app.errorResponse(w, r, http.StatusNotFound, jsonapi.ErrorsPayload{Errors: []*jsonapi.ErrorObject{&errorObject}})
}

func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	m := make(map[string]interface{})
	m["pointer"] = r.RequestURI

	var errorObject = jsonapi.ErrorObject{
		Status: "405",
		Code:   "405",
		Meta:   &m,
		Title:  "Method not allowed",
		Detail: fmt.Sprintf("The %s method is not supported for this resource", r.Method),
	}
	app.errorResponse(w, r, http.StatusMethodNotAllowed, jsonapi.ErrorsPayload{Errors: []*jsonapi.ErrorObject{&errorObject}})
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, source string, err error) {
	m := make(map[string]interface{})
	m["pointer"] = source

	var errorObject = jsonapi.ErrorObject{
		Status: "400",
		Code:   "400",
		Meta:   &m,
		Title:  "Bad request",
		Detail: err.Error(),
	}
	app.errorResponse(w, r, http.StatusMethodNotAllowed, jsonapi.ErrorsPayload{Errors: []*jsonapi.ErrorObject{&errorObject}})
}

func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	var errorsData []*jsonapi.ErrorObject
	for s, s2 := range errors {
		m := make(map[string]interface{})
		m["field"] = s
		errorsData = append(errorsData, &jsonapi.ErrorObject{
			Status: "422",
			Code:   "422",
			Meta:   &m,
			Title:  "Validation failed for field " + s,
			Detail: s2,
		})
	}

	app.errorResponse(w, r, http.StatusUnprocessableEntity, jsonapi.ErrorsPayload{Errors: errorsData})
}

func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request, source string) {
	message := "unable to update the record due to an edit conflict, please try again"
	m := make(map[string]interface{})
	m["pointer"] = source

	var errorObject = jsonapi.ErrorObject{
		Status: "409",
		Code:   "409",
		Meta:   &m,
		Title:  "Data conflict",
		Detail: message,
	}
	app.errorResponse(w, r, http.StatusConflict, jsonapi.ErrorsPayload{Errors: []*jsonapi.ErrorObject{&errorObject}})
}

func (app *application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "Invalid authentication credentials"
	m := make(map[string]interface{})
	m["pointer"] = "createAuthenticationTokenHandler"

	var errorObject = jsonapi.ErrorObject{
		Status: "401",
		Code:   "401",
		Meta:   &m,
		Title:  "Auth Error",
		Detail: message,
	}
	app.errorResponse(w, r, http.StatusUnauthorized, jsonapi.ErrorsPayload{Errors: []*jsonapi.ErrorObject{&errorObject}})
}

func (app *application) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("www-Authenticate", "Bearer")
	message := "Invalid or missing authentication token"
	m := make(map[string]interface{})
	m["pointer"] = "authenticate middleware"

	var errorObject = jsonapi.ErrorObject{
		Status: "401",
		Code:   "401",
		Meta:   &m,
		Title:  "Auth Error",
		Detail: message,
	}
	app.errorResponse(w, r, http.StatusUnauthorized, jsonapi.ErrorsPayload{Errors: []*jsonapi.ErrorObject{&errorObject}})
}

func (app *application) authenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "You must be authenticated to access this resource"
	m := make(map[string]interface{})
	m["pointer"] = "authenticate middleware"

	var errorObject = jsonapi.ErrorObject{
		Status: "401",
		Code:   "401",
		Meta:   &m,
		Title:  "Authentication Error",
		Detail: message,
	}
	app.errorResponse(w, r, http.StatusUnauthorized, jsonapi.ErrorsPayload{Errors: []*jsonapi.ErrorObject{&errorObject}})
}

func (app *application) inactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	message := "Your account must be activated to access this resource"
	m := make(map[string]interface{})
	m["pointer"] = "authenticate middleware"

	var errorObject = jsonapi.ErrorObject{
		Status: "403",
		Code:   "402",
		Meta:   &m,
		Title:  "Auth Error",
		Detail: message,
	}
	app.errorResponse(w, r, http.StatusForbidden, jsonapi.ErrorsPayload{Errors: []*jsonapi.ErrorObject{&errorObject}})
}

func (app *application) notPermittedResponse(w http.ResponseWriter, r *http.Request) {
	message := "Your account does not have necessary permissions for this endpoint"
	m := make(map[string]interface{})
	m["pointer"] = "permissions middleware"

	var errorObject = jsonapi.ErrorObject{
		Status: "403",
		Code:   "403",
		Meta:   &m,
		Title:  "Permissions Error",
		Detail: message,
	}
	app.errorResponse(w, r, http.StatusForbidden, jsonapi.ErrorsPayload{Errors: []*jsonapi.ErrorObject{&errorObject}})
}

func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	m := make(map[string]interface{})
	m["pointer"] = "rateLimit middleware"

	var errorObject = jsonapi.ErrorObject{
		Status: "429",
		Code:   "429",
		Meta:   &m,
		Title:  "Rate Limit error",
		Detail: message,
	}
	app.errorResponse(w, r, http.StatusTooManyRequests, jsonapi.ErrorsPayload{Errors: []*jsonapi.ErrorObject{&errorObject}})
}
