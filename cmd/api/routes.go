package main

import (
	"expvar"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) routes() http.Handler {
	var router = httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/", home)
	router.HandlerFunc(http.MethodGet, "/activate", activation)
	router.HandlerFunc(http.MethodGet, "/public/:id", app.publicList)
	router.HandlerFunc(http.MethodGet, "/reset-password", resetPasswordHandler)
	router.ServeFiles("/static/*filepath", http.Dir("ui/static"))
	router.ServeFiles("/storage/*filepath", http.Dir("storage"))

	fileServer := http.FileServer(http.Dir("./"))

	router.HandlerFunc(http.MethodGet, "/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fileServer.ServeHTTP(w, r)
	})

	router.HandlerFunc(http.MethodGet, "/api/v1/healthcheck", app.healthCheckHandler)

	router.HandlerFunc(http.MethodGet, "/api/v1/folders", app.requirePermission("folders:read", app.indexFoldersHandler))
	router.HandlerFunc(http.MethodPost, "/api/v1/folders", app.requirePermission("folders:write", app.createFolderHandler))
	router.HandlerFunc(http.MethodGet, "/api/v1/folders/:id", app.requirePermission("folders:read", app.showFolderByIdHandler))
	router.HandlerFunc(http.MethodPatch, "/api/v1/folders/:id", app.requirePermission("folders:write", app.updateFolderHandler))
	router.HandlerFunc(http.MethodDelete, "/api/v1/folders/:id", app.requirePermission("folders:write", app.deleteFolderHandler))

	router.HandlerFunc(http.MethodGet, "/api/v1/folders/:id/lists", app.requirePermission("lists:read", app.indexListsHandler))
	router.HandlerFunc(http.MethodGet, "/api/v1/lists", app.requirePermission("lists:read", app.indexListsHandler))
	router.HandlerFunc(http.MethodGet, "/api/v1/lists/:id", app.requirePermission("lists:read", app.showListByIdHandler))
	router.HandlerFunc(http.MethodPost, "/api/v1/lists", app.requirePermission("lists:write", app.createListsHandler))
	router.HandlerFunc(http.MethodPatch, "/api/v1/lists/:id", app.requirePermission("lists:write", app.updateListHandler))
	router.HandlerFunc(http.MethodDelete, "/api/v1/lists/:id", app.requirePermission("lists:write", app.deleteListHandler))

	router.HandlerFunc(http.MethodGet, "/api/v1/links/:link", app.showPublicListHandler)

	router.HandlerFunc(http.MethodGet, "/api/v1/lists/:id/items", app.requirePermission("items:read", app.indexItemsHandler))
	router.HandlerFunc(http.MethodGet, "/api/v1/items", app.requirePermission("items:read", app.indexItemsHandler))
	router.HandlerFunc(http.MethodGet, "/api/v1/items/:id", app.requirePermission("items:read", app.showItemByIdHandler))
	router.HandlerFunc(http.MethodPost, "/api/v1/items", app.requirePermission("items:write", app.createItemsHandler))
	router.HandlerFunc(http.MethodPatch, "/api/v1/items/:id", app.requirePermission("items:write", app.updateItemHandler))
	router.HandlerFunc(http.MethodDelete, "/api/v1/items/:id", app.requirePermission("items:write", app.deleteItemHandler))
	router.HandlerFunc(http.MethodPatch, "/api/v1/lists/:id/items/undone", app.requirePermission("items:write", app.uncrossAllItems))
	router.HandlerFunc(http.MethodDelete, "/api/v1/lists/:id/items", app.requirePermission("items:write", app.deleteAllItemsFromListHandler))
	router.HandlerFunc(http.MethodDelete, "/api/v1/lists/:id/items/done", app.requirePermission("items:write", app.deleteDoneItemsFromListHandler))
	router.HandlerFunc(http.MethodPost, "/api/v1/lists/:id/email", app.requirePermission("items:read", app.sendListByEmail))

	router.HandlerFunc(http.MethodPost, "/api/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPatch, "/api/v1/users/:id", app.updateUserHandler)
	router.HandlerFunc(http.MethodGet, "/api/v1/my", app.showCurrentUserHandler)
	router.HandlerFunc(http.MethodPut, "/api/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPut, "/api/v1/users/password", app.resetUserPasswordHandler)
	router.HandlerFunc(http.MethodDelete, "/api/v1/users/:id", app.deleteUserHandler)

	router.HandlerFunc(http.MethodPost, "/api/v1/tokens/authentication", app.createAuthenticationTokenHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/tokens/password-reset", app.createPasswordResetTokenHandler)

	router.Handler(http.MethodGet, "/api/v1/debug/vars", expvar.Handler())

	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
