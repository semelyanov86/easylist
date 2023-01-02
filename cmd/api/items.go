package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

func (app *application) createItemsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(w, "create a new item")
}

func (app *application) showItemsHandler(w http.ResponseWriter, r *http.Request) {
	var params = httprouter.ParamsFromContext(r.Context())
	listId, err := strconv.ParseInt(params.ByName("list"), 10, 64)
	if err != nil || listId < 1 {
		http.NotFound(w, r)
		return
	}
	_, err = fmt.Fprintf(w, "showing all items of list %d\n", listId)
	if err != nil {
		panic(err)
	}
}
