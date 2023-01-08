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
	type attributes struct {
		Name         string  `json:"name"`
		Description  string  `json:"description"`
		Quantity     int32   `json:"quantity"`
		QuantityType string  `json:"quantity_type"`
		Price        float32 `json:"price"`
		IsStarred    bool    `json:"is_starred"`
		ListId       int64   `json:"list_id"`
		Order        int32   `json:"order"`
		File         string  `json:"file"`
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
		app.badRequestResponse(w, r, "createItemsHandler", err)
		return
	}
	_, err = app.models.Lists.Get(input.Data.Attributes.ListId, userModel.ID)
	var v = validator.New()
	if err != nil {
		v.AddError("data.attributes.list_id", "Can not find current list id")
	}

	v.Check(input.Data.Type == "items", "data.type", "Wrong type provided, accepted type is items")

	var item = &data.Item{
		Name:         input.Data.Attributes.Name,
		Description:  input.Data.Attributes.Description,
		ListId:       input.Data.Attributes.ListId,
		Price:        input.Data.Attributes.Price,
		IsStarred:    input.Data.Attributes.IsStarred,
		Quantity:     input.Data.Attributes.Quantity,
		QuantityType: input.Data.Attributes.QuantityType,
		Version:      1,
		Order:        1,
		UserId:       userModel.ID,
		CreatedAt:    time.Time{},
		UpdatedAt:    time.Time{},
	}
	if data.ValidateItem(v, item); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	fileName, err := app.saveFile(input.Data.Attributes.File, userModel.ID)
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

	err = app.writeJSON(w, http.StatusCreated, envelope{
		Id:         item.ID,
		TypeData:   ItemType,
		Attributes: item,
	}, headers)

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

	var envelope = envelope{
		Id:         id,
		TypeData:   ItemType,
		Attributes: item,
	}
	err = app.writeJSON(w, http.StatusOK, envelope, nil)
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

	type attributes struct {
		ListId       *int64   `json:"list_id"`
		Name         *string  `json:"name"`
		Description  *string  `json:"description"`
		Quantity     *int32   `json:"quantity"`
		QuantityType *string  `json:"quantity_type"`
		Price        *float32 `json:"price"`
		IsStarred    *bool    `json:"is_starred"`
		File         *string  `json:"file"`
		Order        *int32   `json:"order"`
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

	v.Check(input.Data.Type == ItemType, "data.type", "Wrong type provided, accepted type is items")
	v.Check(input.Data.Id == strconv.FormatInt(id, 10), "data.id", "Passed json id does not match request id")
	if err != nil {
		app.badRequestResponse(w, r, "updateItemHandler", err)
		return
	}

	if input.Data.Attributes.Name != nil {
		item.Name = *input.Data.Attributes.Name
	}
	if input.Data.Attributes.ListId != nil {
		item.ListId = *input.Data.Attributes.ListId
	}
	if input.Data.Attributes.Description != nil {
		item.Description = *input.Data.Attributes.Description
	}
	if input.Data.Attributes.Quantity != nil {
		item.Quantity = *input.Data.Attributes.Quantity
	}
	if input.Data.Attributes.QuantityType != nil {
		item.QuantityType = *input.Data.Attributes.QuantityType
	}
	if input.Data.Attributes.Price != nil {
		item.Price = *input.Data.Attributes.Price
	}
	if input.Data.Attributes.IsStarred != nil {
		item.IsStarred = *input.Data.Attributes.IsStarred
	}
	if input.Data.Attributes.Order != nil {
		item.Order = *input.Data.Attributes.Order
	}
	if input.Data.Attributes.File != nil {
		fileName, err := app.saveFile(*input.Data.Attributes.File, userModel.ID)
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

	err = app.writeJSON(w, http.StatusOK, envelope{
		Id:         item.ID,
		TypeData:   ListType,
		Attributes: item,
	}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
