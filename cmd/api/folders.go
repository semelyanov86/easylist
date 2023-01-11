package main

import (
	"database/sql"
	"easylist/internal/data"
	"easylist/internal/validator"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const FolderType = "folders"

type FolderInput struct {
	Name string
	data.Filters
}

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
		Name:    input.Data.Attributes.Name,
		Icon:    input.Data.Attributes.Icon,
		Version: 1,
		Order:   input.Data.Attributes.Order,
		UserId: sql.NullInt64{
			Int64: userModel.ID,
			Valid: true,
		},
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

func (app *application) indexFoldersHandler(w http.ResponseWriter, r *http.Request) {
	v := validator.New()
	qs := r.URL.Query()
	var input FolderInput
	var userModel = app.contextGetUser(r)
	input.Name = app.readString(qs, "filter[name]", "")
	input.Filters.Page = app.readInt(qs, "page[number]", 1, v)
	input.Filters.Size = app.readInt(qs, "page[size]", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "order")

	input.Filters.SortSafelist = []string{"id", "name", "order", "created_at", "updated_at", "-id", "-name", "-order", "-created_at", "-updated_at"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	folders, err := app.models.Folders.GetAll(input.Name, userModel.ID, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, folders, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showFolderByIdHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	var userModel = app.contextGetUser(r)
	folder, err := app.models.Folders.Get(id, userModel.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, folder, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateFolderHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	var userModel = app.contextGetUser(r)

	folder, err := app.models.Folders.Get(id, userModel.ID)
	var oldOrder = folder.Order

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	type attributes struct {
		Name  *string `json:"name"`
		Icon  *string `json:"icon"`
		Order *int32  `json:"order"`
	}
	type inputAttributes struct {
		Id         string     `json:"id"`
		Type       string     `json:"type"`
		Attributes attributes `json:"attributes"`
	}
	var input struct {
		Data inputAttributes `json:"data"`
	}

	var v = validator.New()

	err = app.readJSON(w, r, &input)

	v.Check(input.Data.Type == FolderType, "data.type", "Wrong type provided, accepted type is folders")
	v.Check(input.Data.Id == strconv.FormatInt(id, 10), "data.id", "Passed json id does not match request id")
	if err != nil {
		app.badRequestResponse(w, r, "updateFolderHandler", err)
		return
	}

	if input.Data.Attributes.Name != nil {
		folder.Name = *input.Data.Attributes.Name
	}
	if input.Data.Attributes.Icon != nil {
		folder.Icon = *input.Data.Attributes.Icon
	}
	if input.Data.Attributes.Order != nil {
		folder.Order = *input.Data.Attributes.Order
	}

	v.Check(folder.Order > 0, "data.attributes.order", "order should be greater then zero")

	if data.ValidateFolder(v, folder); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Folders.Update(folder, oldOrder)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r, "updateFolderHandler")
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if r.Header.Get("X-Expected-Version") != "" {
		if strconv.FormatInt(int64(int32(folder.Version)), 32) != r.Header.Get("X-Expected-Version") {
			app.editConflictResponse(w, r, "updateFolderHandler")
			return
		}
	}

	err = app.writeJSON(w, http.StatusOK, envelope{
		Id:         folder.ID,
		TypeData:   FolderType,
		Attributes: folder,
	}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteFolderHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	var userModel = app.contextGetUser(r)

	err = app.models.Folders.Delete(id, userModel.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
