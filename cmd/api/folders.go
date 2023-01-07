package main

import (
	"easylist/internal/data"
	"easylist/internal/validator"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"time"
)

const FolderType = "folders"

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

	var userModel = app.contextGetUser(r)

	var folder = &data.Folder{
		Name:      input.Data.Attributes.Name,
		Icon:      input.Data.Attributes.Icon,
		Version:   1,
		Order:     input.Data.Attributes.Order,
		UserId:    userModel.ID,
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}
	// v.Check(folder.Order > 0, "data.attributes.order", "order should be greater then zero")
	if data.ValidateFolder(v, folder); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Folders.Insert(folder)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	var headers = make(http.Header)
	headers.Set("Location", fmt.Sprintf("/api/v1/folders/%d", folder.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{
		Id:         folder.ID,
		TypeData:   FolderType,
		Attributes: folder,
	}, headers)

	if err != nil {
		app.serverErrorResponse(w, r, err)
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
		TypeData:   FolderType,
		Attributes: folder,
	}
	err = app.writeJSON(w, http.StatusOK, envelope, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
