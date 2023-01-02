package main

import (
	"fmt"
	"net/http"
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
