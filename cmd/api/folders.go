package main

import (
	"easylist/internal/data"
	"easylist/internal/validator"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"time"
)

func (app *application) createFolderHandler(w http.ResponseWriter, r *http.Request) {
	type attributes struct {
		Name  string `json:"name"`
		Icon  string `json:"icon"`
		Order int32  `json:"order"`
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
		app.badRequestResponse(w, r, "createFolderHandler", err)
		return
	}
	var v = validator.New()
	v.Check(input.Data.Type == "folders", "data.type", "Wrong type provided, accepted type is folders")

	var folder = &data.Folder{
		Name:      input.Data.Attributes.Name,
		Icon:      input.Data.Attributes.Icon,
		Version:   1,
		Order:     input.Data.Attributes.Order,
		UserId:    1,
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}
	if data.ValidateFolder(v, folder); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	_, err = fmt.Fprintf(w, "%+v\n", input)
	if err != nil {
		return
	}
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
