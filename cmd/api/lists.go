package main

import (
	"easylist/internal/data"
	"fmt"
	"net/http"
	"time"
)

func (app *application) createListsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(w, "create a new list")
}

func (app *application) showListsHandler(w http.ResponseWriter, r *http.Request) {
	folderId, err := app.readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_, err = fmt.Fprintf(w, "showing all lists from folder %d\n", folderId)
	if err != nil {
		panic(err)
	}
}

func (app *application) showListByIdHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	var list = data.List{
		ID:        id,
		UserId:    1,
		FolderId:  2,
		Name:      "Knoblauch",
		Icon:      "fa-icon",
		Link:      "rwerewrwe",
		Order:     1,
		Version:   1,
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}
	var envelope = envelope{
		Id:         id,
		TypeData:   "list",
		Attributes: list,
	}
	err = app.writeJSON(w, http.StatusOK, envelope, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
