package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const StoragePath = "storage/"

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

func (app *application) readFieldset(r *http.Request, typeData string) []string {
	var fieldString = r.URL.Query().Get("fields[" + typeData + "]")
	var result []string
	if len(fieldString) < 1 {
		return result
	}
	var fields = strings.Split(fieldString, ",")
	for _, field := range fields {
		if field != "id" {
			result = append(result, field)
		}
	}
	return result
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

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	var maxBytes = 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	var dec = json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var err = dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at characted %d)", unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case strings.HasPrefix(err.Error(), "json:unknown field "):
			var fieldName = strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}

func (app *application) background(fn func()) {
	app.wg.Add(1)

	go func() {
		defer app.wg.Done()
		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()
		fn()
	}()
}

func (app *application) saveFile(file string, userId int64) (string, error) {
	if len(file) > 0 {
		// we have a photo
		decoded, err := base64.StdEncoding.DecodeString(file)
		if err != nil {
			return "", err
		}

		var fileUuid = uuid.NewString()
		var fileName = fmt.Sprintf("%scovers/%d/%s.jpg", StoragePath, userId, fileUuid)
		err = os.MkdirAll(fmt.Sprintf("%scovers/%d/", StoragePath, userId), os.ModePerm)
		if err != nil {
			return "", err
		}
		// write image to /storage/covers
		if err := os.WriteFile(fileName, decoded, 0666); err != nil {
			return "", err
		}
		return fileName, nil
	}
	return "", nil
}
