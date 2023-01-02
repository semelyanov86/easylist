package main

import (
	"encoding/json"
	"errors"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

type envelope struct {
	Id         int64  `json:"id,omitempty,string"`
	TypeData   string `json:"type"`
	Attributes any    `json:"attributes"`
}

type envelopeData struct {
	Data envelope `json:"data"`
}

func (app *application) readIDParam(r *http.Request) (int64, error) {
	var params = httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	var envData = envelopeData{Data: data}
	js, err := json.Marshal(envData)
	if err != nil {
		return err
	}
	js = append(js, '\n')
	for key, value := range headers {
		w.Header()[key] = value
	}
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(status)
	_, err = w.Write(js)
	if err != nil {
		return err
	}
	return nil
}
