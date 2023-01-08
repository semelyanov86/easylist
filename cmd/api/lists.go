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

	var userModel = app.contextGetUser(r)

	var err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, "createListHandler", err)
		return
	}
	var v = validator.New()
	v.Check(input.Data.Type == "lists", "data.type", "Wrong type provided, accepted type is lists")

	var list = &data.List{
		ID:        1,
		UserId:    userModel.ID,
		FolderId:  input.Data.Attributes.FolderId,
		Name:      input.Data.Attributes.Name,
		Icon:      input.Data.Attributes.Icon,
		Order:     1,
		Version:   1,
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}

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

	err = app.writeJSON(w, http.StatusCreated, envelope{
		Id:         list.ID,
		TypeData:   ListType,
		Attributes: list,
	}, headers)

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

	var envelope = envelope{
		Id:         id,
		TypeData:   ListType,
		Attributes: list,
	}
	err = app.writeJSON(w, http.StatusOK, envelope, nil)
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

	type attributes struct {
		Name     *string `json:"name"`
		Icon     *string `json:"icon"`
		Order    *int32  `json:"order"`
		FolderId *int64  `json:"folder_id"`
		IsPublic bool    `json:"is_public"`
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

	v.Check(input.Data.Type == ListType, "data.type", "Wrong type provided, accepted type is lists")
	v.Check(input.Data.Id == strconv.FormatInt(id, 10), "data.id", "Passed json id does not match request id")
	if err != nil {
		app.badRequestResponse(w, r, "updateListHandler", err)
		return
	}

	if input.Data.Attributes.Name != nil {
		list.Name = *input.Data.Attributes.Name
	}
	if input.Data.Attributes.Icon != nil {
		list.Icon = *input.Data.Attributes.Icon
	}
	if input.Data.Attributes.Order != nil {
		list.Order = *input.Data.Attributes.Order
	}
	if input.Data.Attributes.FolderId != nil {
		list.FolderId = *input.Data.Attributes.FolderId
	}
	if input.Data.Attributes.IsPublic && !list.Link.Valid {
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

	err = app.writeJSON(w, http.StatusOK, envelope{
		Id:         list.ID,
		TypeData:   ListType,
		Attributes: list,
	}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
