package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type errorEnvelope struct {
	Errors []errorData `json:"errors"`
}

type errorData struct {
	Status string            `json:"status"`
	Source map[string]string `json:"source"`
	Title  string            `json:"title"`
	Detail string            `json:"detail"`
}

func (app *application) logError(r *http.Request, err error) {
	app.logger.PrintError(err, nil)
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message []errorData) {
	var env = errorEnvelope{
		message,
	}
	js, err := json.Marshal(env)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
	js = append(js, '\n')
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(status)
	_, err = w.Write(js)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)
	var errors []errorData
	app.errorResponse(w, r, http.StatusInternalServerError, append(errors, errorData{
		Status: "500",
		Source: map[string]string{"pointer": string(rune(log.Lshortfile))},
		Title:  "Internal server error",
		Detail: "The server encountered a problem and could not process your request",
	}))
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	var errors []errorData
	app.errorResponse(w, r, http.StatusNotFound, append(errors, errorData{
		Status: "404",
		Source: map[string]string{"pointer": r.RequestURI},
		Title:  "Not Found",
		Detail: "The requested resource could not be found",
	}))
}

func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	var errors []errorData
	app.errorResponse(w, r, http.StatusMethodNotAllowed, append(errors, errorData{
		Status: "405",
		Source: map[string]string{"pointer": r.RequestURI},
		Title:  "Method not allowed",
		Detail: fmt.Sprintf("The %s method is not supported for this resource", r.Method),
	}))
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, source string, err error) {
	var errors []errorData
	app.errorResponse(w, r, http.StatusBadRequest, append(errors, errorData{
		Status: "400",
		Source: map[string]string{"source": source},
		Title:  "Bad request",
		Detail: err.Error(),
	}))
}

func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	var errorsData []errorData
	for s, s2 := range errors {
		errorsData = append(errorsData, errorData{
			Status: "422",
			Source: map[string]string{"field": s},
			Title:  "Validation failed for field " + s,
			Detail: s2,
		})
	}
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errorsData)
}
