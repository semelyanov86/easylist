package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) routes() *httprouter.Router {
	var router = httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/", home)
	router.ServeFiles("/static/*filepath", http.Dir("ui/static"))

	router.HandlerFunc(http.MethodGet, "/api/v1/healthcheck", app.healthCheckHandler)
	router.HandlerFunc(http.MethodGet, "/api/v1/folders", app.showFoldersHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/folders", app.createFolderHandler)
	router.HandlerFunc(http.MethodGet, "/api/v1/folders/:id", app.showFolderByIdHandler)

	router.HandlerFunc(http.MethodGet, "/api/v1/folders/:id/lists", app.showListsHandler)
	router.HandlerFunc(http.MethodGet, "/api/v1/lists/:id", app.showListByIdHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/lists", app.createListsHandler)

	router.HandlerFunc(http.MethodGet, "/api/v1/lists/:id/items", app.showItemsHandler)
	router.HandlerFunc(http.MethodGet, "/api/v1/items/:id", app.showItemByIdHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/items", app.createItemsHandler)
	return router
}
