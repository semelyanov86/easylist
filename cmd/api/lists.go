package main

import (
	"database/sql"
	"easylist/internal/data"
	"easylist/internal/validator"
	"errors"
	"fmt"
	"github.com/google/jsonapi"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
	"time"
)

const ListType = "lists"

type ListInput struct {
	Name string
	data.Filters
}

type EmailData struct {
	Items  data.Items
	List   *data.List
	User   *data.User
	Logo   string
	Domain string
}

func (app *application) createListsHandler(w http.ResponseWriter, r *http.Request) {
	var list = new(data.List)
	if err := readJsonApi(r, list); err != nil {
		app.badRequestResponse(w, r, "createListsHandler", err)
		return
	}

	var userModel = app.contextGetUser(r)

	var folderId = list.FolderId
	if folderId == 0 {
		folderId = 1
	}
	_, err := app.models.Folders.Get(folderId, userModel.ID)
	var v = validator.New()
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("data.attributes.folder_id", "this folder does not exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	list.ID = 1
	list.UserId = userModel.ID
	list.Order = 1
	list.Version = 1
	list.CreatedAt = time.Now()
	list.UpdatedAt = time.Now()

	if data.ValidateList(v, list); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Lists.Insert(list)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	var headers = make(http.Header)
	headers.Set("Location", fmt.Sprintf("%s/api/v1/lists/%d", app.config.Domain, list.ID))

	err = app.writeJSON(w, http.StatusCreated, list, headers)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) indexListsHandler(w http.ResponseWriter, r *http.Request) {
	folderId, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	var v = validator.New()
	var qs = r.URL.Query()
	var input ListInput
	var userModel = app.contextGetUser(r)

	if folderId > 0 {
		_, err = app.models.Folders.Get(folderId, userModel.ID)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.notPermittedResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}
	}

	input.Name = app.readString(qs, "filter[name]", "")
	input.Filters.Page = app.readInt(qs, jsonapi.QueryParamPageNumber, 1, v)
	input.Filters.Size = app.readInt(qs, jsonapi.QueryParamPageSize, 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "order")
	input.Filters.SortSafelist = []string{"id", "name", "order", "created_at", "updated_at", "folder_id", "-id", "-name", "-order", "-created_at", "-updated_at", "-folder_id"}
	input.Filters.Includes = app.readCSV(qs, "include", []string{})

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	lists, metadata, err := app.models.Lists.GetAll(folderId, input.Name, userModel.ID, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = writeAndChangeJson(w, http.StatusOK, lists, metadata, ListType, app.config.Domain)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showListByIdHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	var userModel = app.contextGetUser(r)
	list, err := app.models.Lists.Get(id, userModel.ID)
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
	if len(includes) > 0 && data.Contains(includes, "folder") {
		folderModel, err := app.models.Folders.Get(list.FolderId, userModel.ID)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.notFoundResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}
		list.Folder = folderModel
	}
	if len(includes) > 0 && data.Contains(includes, "items") {
		v := validator.New()
		var input = app.NewItemInput(r, v)

		items, _, err := app.models.Items.GetAll("", userModel.ID, list.ID, false, input.Filters)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		list.Items = items
	}

	err = app.writeJSON(w, http.StatusOK, list, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) sendListByEmail(w http.ResponseWriter, r *http.Request) {
	var emailList EmailData
	emailList.Logo = "https://sergeyem.ru/img/easylist-logo.png"
	emailList.Domain = app.config.Domain
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	var userModel = app.contextGetUser(r)
	emailList.User = userModel

	var emailInput = new(data.EmailInput)
	if err := readJsonApi(r, emailInput); err != nil {
		app.badRequestResponse(w, r, "sendListByEmail", err)
		return
	}

	list, err := app.models.Lists.Get(id, userModel.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	emailList.List = list

	v := validator.New()
	var input = app.NewItemInput(r, v)
	input.Filters.Size = 100
	items, _, err := app.models.Items.GetAll("", userModel.ID, id, false, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	emailList.Items = items
	app.background(func() {
		err = app.mailer.Send(emailInput.Email, "list_email.tmpl", emailList)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) showPublicListHandler(w http.ResponseWriter, r *http.Request) {
	var params = httprouter.ParamsFromContext(r.Context())
	var link = params.ByName("link")

	list, err := app.models.Lists.GetPublic(link)
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
	if len(includes) > 0 && data.Contains(includes, "items") {
		v := validator.New()
		var input = app.NewItemInput(r, v)

		items, _, err := app.models.Items.GetAll("", list.UserId, list.ID, false, input.Filters)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		list.Items = items
	}

	err = app.writeJSON(w, http.StatusOK, list, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateListHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	var userModel = app.contextGetUser(r)

	list, err := app.models.Lists.Get(id, userModel.ID)
	var oldOrder = list.Order

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	_, err = app.models.Folders.Get(list.FolderId, userModel.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notPermittedResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var v = validator.New()

	var inputList = new(data.List)
	if err := readJsonApi(r, inputList); err != nil {
		app.badRequestResponse(w, r, "updateListHandler", err)
		return
	}

	if inputList.Name != "" {
		list.Name = inputList.Name
	}
	if inputList.Icon != "" {
		list.Icon = inputList.Icon
	}
	if inputList.Order != 0 {
		list.Order = inputList.Order
	}
	if inputList.FolderId != 0 {
		list.FolderId = inputList.FolderId
	}

	if inputList.IsPublic == 2 && !list.Link.Valid {
		list.Link = data.Link{
			NullString: sql.NullString{
				String: uuid.NewString(),
				Valid:  true,
			},
		}
	}

	if inputList.IsPublic == 1 && list.Link.Valid {
		list.Link = data.Link{
			NullString: sql.NullString{
				String: "",
				Valid:  false,
			},
		}
	}

	v.Check(list.Order > 0, "data.attributes.order", "order should be greater then zero")

	if data.ValidateList(v, list); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Lists.Update(list, oldOrder)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r, "updateListHandler")
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if r.Header.Get("X-Expected-Version") != "" {
		if strconv.FormatInt(int64(list.Version), 32) != r.Header.Get("X-Expected-Version") {
			app.editConflictResponse(w, r, "updateListHandler")
			return
		}
	}

	err = app.writeJSON(w, http.StatusOK, list, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteListHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	var userModel = app.contextGetUser(r)

	err = app.models.Lists.Delete(id, userModel.ID)
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
