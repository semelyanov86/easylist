package main

import (
	"easylist/internal/data"
	"easylist/internal/validator"
	"fmt"
	"net/http"
	"time"
)

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
		app.badRequestResponse(w, r, "createItemsHandler", err)
		return
	}
	var v = validator.New()
	v.Check(input.Data.Type == "items", "data.type", "Wrong type provided, accepted type is items")

	var item = &data.Item{
		Name:        input.Data.Attributes.Name,
		Description: input.Data.Attributes.Description,
		ListId:      input.Data.Attributes.ListId,
		Price:       input.Data.Attributes.Price,
		IsStarred:   input.Data.Attributes.IsStarred,
		Version:     1,
		Order:       input.Data.Attributes.Order,
		UserId:      1,
		CreatedAt:   time.Time{},
		UpdatedAt:   time.Time{},
	}
	if data.ValidateItem(v, item); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	_, err = fmt.Fprintf(w, "%+v\n", input)
	if err != nil {
		return
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
