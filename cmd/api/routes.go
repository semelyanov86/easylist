package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) routes() *httprouter.Router {
	var router = httprouter.New()
	router.HandlerFunc(http.MethodGet, "/api/v1/healthcheck", app.healthCheckHandler)
	return router
}
