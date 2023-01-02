package main

import (
	"easylist/internal/data"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"time"
)

func (app *application) createFolderHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(w, "create a new folder")
}

func (app *application) showFoldersHandler(w http.ResponseWriter, r *http.Request) {
	_ = httprouter.ParamsFromContext(r.Context())
	_, err := fmt.Fprintf(w, "showing all folders")
	if err != nil {
		panic(err)
	}
}

func (app *application) showFolderByIdHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	var folder = data.Folder{
		ID:        id,
		CreatedAt: time.Now(),
		Name:      "Main folder",
		Icon:      "fa-folder",
		Version:   1,
		Order:     1,
		UserId:    1,
		UpdatedAt: time.Now(),
	}
	var envelope = envelope{
		Id:         id,
		TypeData:   "folder",
		Attributes: folder,
	}
	err = app.writeJSON(w, http.StatusOK, envelope, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
