package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) routes() http.Handler {
	var router = httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/", home)
	router.HandlerFunc(http.MethodGet, "/activate", activation)
	router.ServeFiles("/static/*filepath", http.Dir("ui/static"))

	router.HandlerFunc(http.MethodGet, "/api/v1/healthcheck", app.healthCheckHandler)

	router.HandlerFunc(http.MethodGet, "/api/v1/folders", app.requirePermission("folders:read", app.showFoldersHandler))
	router.HandlerFunc(http.MethodPost, "/api/v1/folders", app.requirePermission("folders:write", app.createFolderHandler))
	router.HandlerFunc(http.MethodGet, "/api/v1/folders/:id", app.requirePermission("folders:read", app.showFolderByIdHandler))
	router.HandlerFunc(http.MethodPatch, "/api/v1/folders/:id", app.requirePermission("folders:write", app.updateFolderHandler))
	router.HandlerFunc(http.MethodDelete, "/api/v1/folders/:id", app.requirePermission("folders:write", app.deleteFolderHandler))

	router.HandlerFunc(http.MethodGet, "/api/v1/folders/:id/lists", app.requirePermission("lists:read", app.showListsHandler))
	router.HandlerFunc(http.MethodGet, "/api/v1/lists/:id", app.requirePermission("lists:read", app.showListByIdHandler))
	router.HandlerFunc(http.MethodPost, "/api/v1/lists", app.requirePermission("lists:write", app.createListsHandler))
	router.HandlerFunc(http.MethodPatch, "/api/v1/lists/:id", app.requirePermission("lists:write", app.updateListHandler))
	router.HandlerFunc(http.MethodDelete, "/api/v1/lists/:id", app.requirePermission("lists:write", app.deleteListHandler))

	router.HandlerFunc(http.MethodGet, "/api/v1/lists/:id/items", app.requirePermission("items:read", app.showItemsHandler))
	router.HandlerFunc(http.MethodGet, "/api/v1/items/:id", app.requirePermission("items:read", app.showItemByIdHandler))
	router.HandlerFunc(http.MethodPost, "/api/v1/items", app.requirePermission("items:write", app.createItemsHandler))

	router.HandlerFunc(http.MethodPost, "/api/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/api/v1/users/activated", app.activateUserHandler)

	router.HandlerFunc(http.MethodPost, "/api/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	return app.recoverPanic(app.authenticate(router))
}
