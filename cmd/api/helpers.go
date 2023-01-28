package main

import (
	"easylist/internal/data"
	"easylist/internal/validator"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/jsonapi"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"io"
	"net/http"
	"net/url"
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

func (app *application) readIDParam(r *http.Request) (int64, error) {
	var params = httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		if id == 0 {
			return 0, nil
		}
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	app.writeHeaders(w, status, headers)
	if err := jsonapi.MarshalPayload(w, data); err != nil {
		return err
	}
	return nil
}

func (app *application) writeHeaders(w http.ResponseWriter, status int, headers http.Header) {
	w.Header().Set("Content-Type", jsonapi.MediaType)
	for key, value := range headers {
		w.Header()[key] = value
	}
	w.WriteHeader(status)
}

func (app *application) writeAndChangeJson(w http.ResponseWriter, status int, data any, metadata data.Metadata, typeData string) error {
	app.writeHeaders(w, status, nil)
	res, err := jsonapi.Marshal(data)
	if err != nil {
		return err
	}

	var relationPrefix string
	if manyPayload, ok := res.(*jsonapi.ManyPayload); ok {
		if metadata.ParentId != 0 {
			relationPrefix = metadata.ParentName + "/" + strconv.Itoa(int(metadata.ParentId)) + "/"
		}
		manyPayload.Links = &jsonapi.Links{
			jsonapi.KeyFirstPage:    app.config.Domain + "/api/v1/" + relationPrefix + typeData + "?" + jsonapi.QueryParamPageNumber + "=" + strconv.Itoa(metadata.FirstPage) + "&" + jsonapi.QueryParamPageSize + "=" + strconv.Itoa(metadata.PageSize),
			jsonapi.KeyPreviousPage: app.config.Domain + "/api/v1/" + relationPrefix + typeData + "?" + jsonapi.QueryParamPageNumber + "=" + strconv.Itoa(metadata.PrevPage) + "&" + jsonapi.QueryParamPageSize + "=" + strconv.Itoa(metadata.PageSize),
			jsonapi.KeyNextPage:     app.config.Domain + "/api/v1/" + relationPrefix + typeData + "?" + jsonapi.QueryParamPageNumber + "=" + strconv.Itoa(metadata.NextPage) + "&" + jsonapi.QueryParamPageSize + "=" + strconv.Itoa(metadata.PageSize),
			jsonapi.KeyLastPage:     app.config.Domain + "/api/v1/" + relationPrefix + typeData + "?" + jsonapi.QueryParamPageNumber + "=" + strconv.Itoa(metadata.LastPage) + "&" + jsonapi.QueryParamPageSize + "=" + strconv.Itoa(metadata.PageSize),
		}
		if metadata.NextPage == 0 {
			(*manyPayload.Links)[jsonapi.KeyNextPage] = nil
		}
		if metadata.PrevPage == 0 {
			(*manyPayload.Links)[jsonapi.KeyPreviousPage] = nil
		}
		(*manyPayload.Meta)["total"] = metadata.TotalRecords
		return json.NewEncoder(w).Encode(manyPayload)
	}
	return jsonapi.ErrInvalidType
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

func (app *application) readJsonApi(r *http.Request, dst any) error {
	return jsonapi.UnmarshalPayload(r.Body, dst)
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

func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	var s = qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return s
}

func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	var csv = qs.Get(key)

	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	var s = qs.Get(key)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return i
}

func (app *application) readBool(qs url.Values, key string, defaultValue bool, v *validator.Validator) bool {
	var s = qs.Get(key)

	if s == "" {
		return defaultValue
	}
	switch s {
	case "false":
		return false
	case "true":
		return true
	case "1":
		return true
	case "0":
		return false
	}

	v.AddError(key, "must be an integer value")
	return defaultValue
}
