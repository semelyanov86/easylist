package main

import (
	"easylist/internal/data"
	"easylist/internal/validator"
	"errors"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"text/template"
)

func (app *application) publicList(w http.ResponseWriter, r *http.Request) {
	var params = httprouter.ParamsFromContext(r.Context())
	var emailData EmailData
	link := params.ByName("id")
	listModel, err := app.models.Lists.GetPublic(link)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	emailData.List = listModel
	v := validator.New()
	var input = app.NewItemInput(r, v)
	input.Filters.Size = 100
	items, _, err := app.models.Items.GetAll("", listModel.UserId, listModel.ID, false, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	emailData.Items = items
	ts, err := template.ParseFiles("./ui/html/public.page.html")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal server error", 500)
		return
	}
	err = ts.Execute(w, emailData)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal server error", 500)
	}
}
