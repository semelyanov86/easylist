package main

import (
	"easylist/internal/data"
	"easylist/internal/validator"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const ItemType = "items"

func (app *application) createItemsHandler(w http.ResponseWriter, r *http.Request) {
	var item = new(data.Item)
	if err := app.readJsonApi(r, item); err != nil {
		app.badRequestResponse(w, r, "createItemsHandler", err)
		return
	}

	var userModel = app.contextGetUser(r)

	_, err := app.models.Lists.Get(item.ListId, userModel.ID)
	var v = validator.New()
	if err != nil {
		v.AddError("data.attributes.list_id", "Can not find current list id")
	}

	item.Version = 1
	item.Order = 1
	item.UserId = userModel.ID
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()

	if data.ValidateItem(v, item); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	fileName, err := app.saveFile(item.File, userModel.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	item.File = fileName

	err = app.models.Items.Insert(item)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	var headers = make(http.Header)
	headers.Set("Location", fmt.Sprintf("/api/v1/items/%d", item.ID))

	err = app.writeJSON(w, http.StatusCreated, item, headers)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showItemsHandler(w http.ResponseWriter, r *http.Request) {
	listId, err := app.readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_, err = fmt.Fprintf(w, "showing all items of list %d\n", listId)
	if err != nil {
		panic(err)
	}
}

func (app *application) showItemByIdHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	var userModel = app.contextGetUser(r)
	item, err := app.models.Items.Get(id, userModel.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, item, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	var userModel = app.contextGetUser(r)

	item, err := app.models.Items.Get(id, userModel.ID)
	var oldOrder = item.Order

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	_, err = app.models.Lists.Get(item.ListId, userModel.ID)
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

	var inputItem = new(data.Item)
	if err := app.readJsonApi(r, inputItem); err != nil {
		app.badRequestResponse(w, r, "updateItemHandler", err)
		return
	}

	if err != nil {
		app.badRequestResponse(w, r, "updateItemHandler", err)
		return
	}

	if inputItem.Name != "" {
		item.Name = inputItem.Name
	}
	if inputItem.ListId != 0 {
		item.ListId = inputItem.ListId
	}
	if inputItem.Description != "" {
		item.Description = inputItem.Description
	}
	if inputItem.Quantity != 0 {
		item.Quantity = inputItem.Quantity
	}
	if inputItem.QuantityType != "" {
		item.QuantityType = inputItem.QuantityType
	}
	if inputItem.Price != 0 {
		item.Price = inputItem.Price
	}
	item.IsStarred = inputItem.IsStarred
	if inputItem.Order != 0 {
		item.Order = inputItem.Order
	}
	if inputItem.File != "" {
		fileName, err := app.saveFile(inputItem.File, userModel.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		item.File = fileName
	}

	v.Check(item.Order > 0, "data.attributes.order", "order should be greater then zero")

	if data.ValidateItem(v, item); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Items.Update(item, oldOrder)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r, "updateItemHandler")
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if r.Header.Get("X-Expected-Version") != "" {
		if strconv.FormatInt(int64(item.Version), 32) != r.Header.Get("X-Expected-Version") {
			app.editConflictResponse(w, r, "updateItemHandler")
			return
		}
	}

	err = app.writeJSON(w, http.StatusOK, item, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	var userModel = app.contextGetUser(r)

	err = app.models.Items.Delete(id, userModel.ID)
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
