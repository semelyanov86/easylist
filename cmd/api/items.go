package main

import (
	"easylist/internal/data"
	"easylist/internal/validator"
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"os"
	"time"
)

const ItemType = "items"
const StoragePath = "storage/"

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

	var fileUuid string
	if len(input.Data.Attributes.File) > 0 {
		// we have a photo
		decoded, err := base64.StdEncoding.DecodeString(input.Data.Attributes.File)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		fileUuid = uuid.NewString()
		var fileName = fmt.Sprintf("%scovers/%d/%s.jpg", StoragePath, userModel.ID, fileUuid)
		err = os.MkdirAll(fmt.Sprintf("%scovers/%d/", StoragePath, userModel.ID), os.ModePerm)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		// write image to /storage/covers
		if err := os.WriteFile(fileName, decoded, 0666); err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		item.File = fileName
	}

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
	var item = data.Item{
		ID:           id,
		UserId:       1,
		ListId:       2,
		Name:         "Super honig",
		Description:  "This is awesome!",
		Quantity:     1,
		QuantityType: "piece",
		Price:        15.6,
		IsStarred:    false,
		File:         "",
		Order:        1,
		Version:      1,
		CreatedAt:    time.Time{},
		UpdatedAt:    time.Time{},
	}
	var envelope = envelope{
		Id:         id,
		TypeData:   "item",
		Attributes: item,
	}
	err = app.writeJSON(w, http.StatusOK, envelope, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
