package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) createFolderHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(w, "create a new folder")
}

func (app *application) showFoldersHandler(w http.ResponseWriter, r *http.Request) {
	_ = httprouter.ParamsFromContext(r.Context())
	fmt.Println(w, "showing all folders")
}
