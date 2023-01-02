package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) routes() *httprouter.Router {
	var router = httprouter.New()
	router.HandlerFunc(http.MethodGet, "/api/v1/healthcheck", app.healthCheckHandler)
	router.HandlerFunc(http.MethodGet, "/api/v1/folders", app.showFoldersHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/folders", app.createFolderHandler)
	router.HandlerFunc(http.MethodGet, "/api/v1/folders/:id/lists", app.showListsHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/lists", app.createListsHandler)
	router.HandlerFunc(http.MethodGet, "/api/v1/lists/:id/items", app.showItemsHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/items", app.createItemsHandler)
	return router
}
