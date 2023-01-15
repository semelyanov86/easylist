package main

import (
	"database/sql"
	"easylist/internal/data"
	"easylist/internal/validator"
	"errors"
	"fmt"
	"github.com/google/jsonapi"
	"net/http"
	"strconv"
	"time"
)

type FolderInput struct {
	Name string
	data.Filters
}

func (app *application) createFolderHandler(w http.ResponseWriter, r *http.Request) {
	var folder = new(data.Folder)
	if err := app.readJsonApi(r, folder); err != nil {
		app.badRequestResponse(w, r, "createFolderHandler", err)
		return
	}

	var v = validator.New()

	var userModel = app.contextGetUser(r)

	folder.UserId = sql.NullInt64{
		Int64: userModel.ID,
		Valid: true,
	}
	folder.Version = 1
	folder.CreatedAt = time.Time{}
	folder.UpdatedAt = time.Time{}

	// v.Check(folder.Order > 0, "data.attributes.order", "order should be greater then zero")
	if data.ValidateFolder(v, folder); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	var err = app.models.Folders.Insert(folder)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	var headers = make(http.Header)
	headers.Set("Location", fmt.Sprintf("/api/v1/folders/%d", folder.ID))

	err = app.writeJSON(w, http.StatusCreated, folder, headers)

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
	input.Filters.Page = app.readInt(qs, jsonapi.QueryParamPageNumber, 1, v)
	input.Filters.Size = app.readInt(qs, jsonapi.QueryParamPageSize, 20, v)
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

	var inputFolder = new(data.Folder)
	if err := app.readJsonApi(r, inputFolder); err != nil {
		app.badRequestResponse(w, r, "updateFolderHandler", err)
		return
	}

	if inputFolder.Name != "" {
		folder.Name = inputFolder.Name
	}
	if inputFolder.Icon != "" {
		folder.Icon = inputFolder.Icon
	}
	if inputFolder.Order != 0 {
		folder.Order = inputFolder.Order
	}

	var v = validator.New()
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
		if strconv.FormatInt(int64(folder.Version), 32) != r.Header.Get("X-Expected-Version") {
			app.editConflictResponse(w, r, "updateFolderHandler")
			return
		}
	}

	err = app.writeJSON(w, http.StatusOK, folder, nil)
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
