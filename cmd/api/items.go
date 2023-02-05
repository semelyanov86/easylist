package main

import (
	"easylist/internal/data"
	"easylist/internal/validator"
	"errors"
	"fmt"
	"github.com/google/jsonapi"
	"net/http"
	"strconv"
	"time"
)

const ItemType = "items"

type ItemInput struct {
	Name string
	data.Filters
}

func (app *application) createItemsHandler(w http.ResponseWriter, r *http.Request) {
	var item = new(data.Item)
	if err := readJsonApi(r, item); err != nil {
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

func (app *application) indexItemsHandler(w http.ResponseWriter, r *http.Request) {
	listId, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	v := validator.New()
	qs := r.URL.Query()
	var input ItemInput
	var userModel = app.contextGetUser(r)
	input.Name = app.readString(qs, "filter[name]", "")
	var isStarred = app.readBool(qs, "filter[is_starred]", false, v)
	input.Filters.Page = app.readInt(qs, jsonapi.QueryParamPageNumber, 1, v)
	input.Filters.Size = app.readInt(qs, jsonapi.QueryParamPageSize, 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "order")
	input.Filters.Includes = app.readCSV(qs, "include", []string{})

	input.Filters.SortSafelist = []string{"id", "name", "order", "created_at", "updated_at", "quantity", "is_starred", "-id", "-name", "-order", "-created_at", "-updated_at", "-quantity", "-is_starred"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	items, metadata, err := app.models.Items.GetAll(input.Name, userModel.ID, listId, isStarred, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = writeAndChangeJson(w, http.StatusOK, items, metadata, data.ItemsType, app.config.Domain)
	if err != nil {
		app.serverErrorResponse(w, r, err)
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

	var qs = r.URL.Query()
	var includes = app.readCSV(qs, "include", []string{})
	if len(includes) > 0 && data.Contains(includes, "list") {
		listModel, err := app.models.Lists.Get(item.ListId, userModel.ID)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.notFoundResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}
		item.List = listModel
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
	var input = Input[ItemAttributes]{Data: InputAttributes[ItemAttributes]{
		Type:       "tokens",
		Attributes: ItemAttributes{},
	}}

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
