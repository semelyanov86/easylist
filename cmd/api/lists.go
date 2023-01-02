package main

import (
	"fmt"
	"net/http"
)

func (app *application) createListsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(w, "create a new list")
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
