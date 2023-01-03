package main

import (
	"easylist/internal/data"
	"easylist/internal/validator"
	"fmt"
	"net/http"
	"time"
)

func (app *application) createListsHandler(w http.ResponseWriter, r *http.Request) {
	type attributes struct {
		Name     string `json:"name"`
		Icon     string `json:"icon"`
		FolderId int64  `json:"folder_id"`
		Order    int32  `json:"order"`
	}

	type inputAttributes struct {
		Type       string     `json:"type"`
		Attributes attributes `json:"attributes"`
	}
	var input struct {
		Data inputAttributes
	}

	var err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, "createListHandler", err)
		return
	}
	var v = validator.New()
	v.Check(input.Data.Type == "lists", "data.type", "Wrong type provided, accepted type is lists")

	var list = &data.List{
		ID:        1,
		UserId:    1,
		FolderId:  input.Data.Attributes.FolderId,
		Name:      input.Data.Attributes.Name,
		Icon:      input.Data.Attributes.Icon,
		Link:      "",
		Order:     input.Data.Attributes.Order,
		Version:   1,
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}
	if data.ValidateList(v, list); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	_, err = fmt.Fprintf(w, "%+v\n", input)
	if err != nil {
		return
	}
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
