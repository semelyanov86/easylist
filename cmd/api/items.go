package main

import (
	"easylist/internal/data"
	"fmt"
	"net/http"
	"time"
)

func (app *application) createItemsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(w, "create a new item")
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
