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
	app.logger.Println(err)
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message errorData) {
	var env = errorEnvelope{[]errorData{
		message,
	}}
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
	app.errorResponse(w, r, http.StatusInternalServerError, errorData{
		Status: "500",
		Source: map[string]string{"pointer": string(rune(log.Lshortfile))},
		Title:  "Internal server error",
		Detail: "The server encountered a problem and could not process your request",
	})
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusNotFound, errorData{
		Status: "404",
		Source: map[string]string{"pointer": r.RequestURI},
		Title:  "Not Found",
		Detail: "The requested resource could not be found",
	})
}

func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusMethodNotAllowed, errorData{
		Status: "405",
		Source: map[string]string{"pointer": r.RequestURI},
		Title:  "Method not allowed",
		Detail: fmt.Sprintf("The %s method is not supported for this resource", r.Method),
	})
}
