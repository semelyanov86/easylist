package main

import (
	"database/sql"
	"easylist/internal/data"
	"easylist/internal/validator"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"strconv"
	"time"
)

const ListType = "lists"

func (app *application) createListsHandler(w http.ResponseWriter, r *http.Request) {
	var list = new(data.List)
	if err := app.readJsonApi(r, list); err != nil {
		app.badRequestResponse(w, r, "createListsHandler", err)
		return
	}

	var userModel = app.contextGetUser(r)

	var folderId = list.FolderId
	if folderId == 0 {
		folderId = 1
	}
	_, err := app.models.Folders.Get(folderId, userModel.ID)
	var v = validator.New()
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("data.attributes.folder_id", "this folder does not exists")
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	list.ID = 1
	list.UserId = userModel.ID
	list.Order = 1
	list.Version = 1
	list.CreatedAt = time.Now()
	list.UpdatedAt = time.Now()

	if data.ValidateList(v, list); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Lists.Insert(list)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	var headers = make(http.Header)
	headers.Set("Location", fmt.Sprintf("/api/v1/lists/%d", list.ID))

	err = app.writeJSON(w, http.StatusCreated, list, headers)

	if err != nil {
		app.serverErrorResponse(w, r, err)
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
	var userModel = app.contextGetUser(r)
	list, err := app.models.Lists.Get(id, userModel.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, list, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateListHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	var userModel = app.contextGetUser(r)

	list, err := app.models.Lists.Get(id, userModel.ID)
	var oldOrder = list.Order

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	_, err = app.models.Folders.Get(list.FolderId, userModel.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notPermittedResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var v = validator.New()

	var inputList = new(data.List)
	if err := app.readJsonApi(r, inputList); err != nil {
		app.badRequestResponse(w, r, "updateListHandler", err)
		return
	}

	if inputList.Name != "" {
		list.Name = inputList.Name
	}
	if inputList.Icon != "" {
		list.Icon = inputList.Icon
	}
	if inputList.Order != 0 {
		list.Order = inputList.Order
	}
	if inputList.FolderId != 0 {
		list.FolderId = inputList.FolderId
	}

	if inputList.IsPublic && !list.Link.Valid {
		list.Link = data.Link{
			NullString: sql.NullString{
				String: uuid.NewString(),
				Valid:  true,
			},
		}
	}

	v.Check(list.Order > 0, "data.attributes.order", "order should be greater then zero")

	if data.ValidateList(v, list); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Lists.Update(list, oldOrder)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r, "updateListHandler")
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if r.Header.Get("X-Expected-Version") != "" {
		if strconv.FormatInt(int64(list.Version), 32) != r.Header.Get("X-Expected-Version") {
			app.editConflictResponse(w, r, "updateListHandler")
			return
		}
	}

	err = app.writeJSON(w, http.StatusOK, list, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteListHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	var userModel = app.contextGetUser(r)

	err = app.models.Lists.Delete(id, userModel.ID)
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
